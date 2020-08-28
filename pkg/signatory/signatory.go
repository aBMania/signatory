package signatory

import (
	"context"
	"encoding/hex"
	stderr "errors"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/ecadlabs/signatory/pkg/config"
	"github.com/ecadlabs/signatory/pkg/cryptoutils"
	"github.com/ecadlabs/signatory/pkg/errors"
	"github.com/ecadlabs/signatory/pkg/tezos"
	"github.com/ecadlabs/signatory/pkg/vault"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrVaultNotFound error return when a vault is not found
	ErrVaultNotFound = errors.Wrap(stderr.New("This key not found in any vault"), http.StatusNotFound)
	// ErrNotSafeToSign error returned when an operation is a potential duplicate
	ErrNotSafeToSign = errors.Wrap(stderr.New("Not safe to sign"), http.StatusForbidden)
)

const (
	logPKH       = "pkh"
	logVault     = "vault"
	logVaultName = "vault_name"
	logOp        = "op"
	logKind      = "kind"
	logKeyID     = "key_id"
	logChainID   = "chain_id"
	logLevel     = "lvl"
	logClient    = "client_pkh"
)

// SignInterceptor is an observer function for signing request
type SignInterceptor func(opt *SignInterceptorOptions, sing func() error) error

// SignInterceptorOptions contains SignInterceptor arguments to avoid confusion
type SignInterceptorOptions struct {
	Address string
	Vault   string
	Op      string
	Kind    []string
}

type Policy struct {
	AllowedOperations []string `yaml:"allowed_operations"`
	AllowedKinds      []string `yaml:"allowed_kinds"`
	LogPayloads       bool     `yaml:"log_payloads"`
	AuthorizedKeys    []string `yaml:"authorized_keys"`
}

// PublicKey contains base58 encoded public key with its hash
type PublicKey struct {
	PublicKey     string
	PublicKeyHash string
	VaultName     string
	ID            string
	Policy        *Policy
}

// Signatory is a struct coordinate signatory action and select vault according to the key being used
type Signatory struct {
	config Config
	vaults map[string]vault.Vault
	cache  keyCache
}

// SignRequest represents a sing request which may be authenticated with the client key
type SignRequest struct {
	ClientPublicKeyHash string // optional, see policy
	PublicKeyHash       string
	Message             []byte
}

type keyVaultPair struct {
	pkh   string
	key   vault.StoredKey
	vault vault.Vault
	name  string
}

type keyCache struct {
	cache map[string]*keyVaultPair
	mtx   sync.Mutex
}

func (k *keyCache) push(pkh string, pair *keyVaultPair) {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	if k.cache == nil {
		k.cache = make(map[string]*keyVaultPair)
	}
	k.cache[pkh] = pair
}

func (k *keyCache) get(pkh string) *keyVaultPair {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	if pair, ok := k.cache[pkh]; ok {
		return pair
	}

	return nil
}

func (k *keyCache) drop() {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	k.cache = nil
}

func (s *Signatory) logger() log.FieldLogger {
	if s.config.Logger != nil {
		return s.config.Logger
	}
	return log.StandardLogger()
}

var defaultPolicy = Policy{
	AllowedOperations: []string{"block", "endorsement"},
}

func (s *Signatory) fetchPolicyOrDefault(keyHash string) *Policy {
	val, ok := s.config.Policy[keyHash]
	if !ok {
		return nil
	}
	if val != nil {
		return val
	}
	return &defaultPolicy
}

func matchFilter(policy *Policy, req *SignRequest, msg tezos.UnsignedMessage) error {
	if policy.AuthorizedKeys != nil {
		if req.ClientPublicKeyHash == "" {
			return errors.New("authentication required")
		}

		var allowed bool
		for _, k := range policy.AuthorizedKeys {
			if k == req.ClientPublicKeyHash {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("client `%s' is not allowed", req.ClientPublicKeyHash)
		}
	}

	kind := msg.MessageKind()
	var allowed bool
	for _, k := range policy.AllowedOperations {
		if kind == k {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("request kind `%s' is not allowed", kind)
	}

	if ops, ok := msg.(*tezos.UnsignedOperation); ok {
		for _, op := range ops.Contents {
			kind := op.OperationKind()
			allowed = false
			for _, k := range policy.AllowedKinds {
				if kind == k {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("operation `%s' is not allowed", kind)
			}
		}
	}
	return nil
}

// Sign ask the vault to sign a message with the private key associated to keyHash
func (s *Signatory) Sign(ctx context.Context, req *SignRequest) (string, error) {
	l := s.logger().WithField(logPKH, req.PublicKeyHash)

	if req.ClientPublicKeyHash != "" {
		l = l.WithField(logClient, req.ClientPublicKeyHash)
	}

	policy := s.fetchPolicyOrDefault(req.PublicKeyHash)
	if policy == nil {
		err := fmt.Errorf("%s is not listed in config", req.PublicKeyHash)
		l.WithField("raw", hex.EncodeToString(req.Message)).Error(err)
		return "", errors.Wrap(err, http.StatusForbidden)
	}

	msg, err := tezos.ParseUnsignedMessage(req.Message)
	if err != nil {
		l.WithField("raw", hex.EncodeToString(req.Message)).Error(err)
		return "", errors.Wrap(err, http.StatusBadRequest)
	}

	l = l.WithField(logOp, msg.MessageKind())

	if m, ok := msg.(tezos.MessageWithChainID); ok {
		l = l.WithField(logChainID, m.GetChainID())
	}

	if m, ok := msg.(tezos.MessageWithLevel); ok {
		l = l.WithField(logLevel, m.GetLevel())
	}

	var opKind []string
	if ops, ok := msg.(*tezos.UnsignedOperation); ok {
		opKind = ops.OperationKinds()
		l = l.WithField(logKind, opKind)
	}

	p, err := s.getPublicKey(ctx, req.PublicKeyHash)
	if err != nil {
		l.Error(err)
		return "", err
	}

	l = l.WithField(logVault, p.vault.Name())
	if n, ok := p.vault.(vault.VaultNamer); ok {
		l = l.WithField(logVaultName, n.VaultName())
	} else {
		l = l.WithField(logVaultName, p.name)
	}

	if err = matchFilter(policy, req, msg); err != nil {
		l.Error(err)
		return "", errors.Wrap(err, http.StatusForbidden)
	}

	l.Info("Requesting signing operation")

	level := log.DebugLevel
	if policy.LogPayloads {
		level = log.InfoLevel
	}
	l.WithField("raw", hex.EncodeToString(req.Message)).Log(level, "About to sign raw bytes")

	if !s.config.Watermark.IsSafeToSign(req.PublicKeyHash, msg) {
		err = ErrNotSafeToSign
		l.Error(err)
		return "", err
	}

	// Not nil if vault found
	digest := tezos.DigestFunc(req.Message)

	var sig cryptoutils.Signature
	if s.config.Interceptor != nil {
		err = s.config.Interceptor(&SignInterceptorOptions{
			Address: req.PublicKeyHash,
			Vault:   p.vault.Name(),
			Op:      msg.MessageKind(),
			Kind:    opKind,
		}, func() (err error) {
			sig, err = p.vault.Sign(ctx, digest[:], p.key)
			return err
		})
	} else {
		sig, err = p.vault.Sign(ctx, digest[:], p.key)
	}
	if err != nil {
		return "", err
	}

	sig = cryptoutils.CanonizeSignature(sig)

	l.WithField("raw", sig).Debug("Signed bytes")

	encodedSig, err := tezos.EncodeSignature(sig)
	if err != nil {
		return "", err
	}

	l.Debugf("Encoded signature: %s", encodedSig)
	l.Infof("Signed %s successfully", msg.MessageKind())

	return encodedSig, nil
}

func (s *Signatory) listPublicKeys(ctx context.Context) (ret map[string]*keyVaultPair, list []*keyVaultPair, err error) {
	ret = make(map[string]*keyVaultPair)
	for name, v := range s.vaults {
		iter := v.ListPublicKeys(ctx)
		for {
			key, err := iter.Next()
			if err == vault.ErrDone {
				break
			}
			if err != nil {
				return nil, nil, err
			}

			pkh, err := tezos.EncodePublicKeyHash(key.PublicKey())
			if err != nil {
				return nil, nil, err
			}
			p := &keyVaultPair{pkh: pkh, key: key, vault: v, name: name}
			s.cache.push(pkh, p)

			ret[pkh] = p
			list = append(list, p)
		}
	}
	return ret, list, nil
}

// ListPublicKeys retrieve the list of all public keys supported by the current configuration
func (s *Signatory) ListPublicKeys(ctx context.Context) ([]*PublicKey, error) {
	_, list, err := s.listPublicKeys(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]*PublicKey, len(list))
	for i, p := range list {
		enc, err := tezos.EncodePublicKey(p.key.PublicKey())
		if err != nil {
			return nil, err
		}
		ret[i] = &PublicKey{
			PublicKey:     enc,
			PublicKeyHash: p.pkh,
			VaultName:     p.vault.Name(),
			ID:            p.key.ID(),
			Policy:        s.fetchPolicyOrDefault(p.pkh),
		}
	}
	return ret, nil
}

func (s *Signatory) getPublicKey(ctx context.Context, keyHash string) (*keyVaultPair, error) {
	cached := s.cache.get(keyHash)
	if cached != nil {
		return cached, nil
	}

	s.logger().WithField(logPKH, keyHash).Debugf("Fetching public key for: %s", keyHash)

	keys, _, err := s.listPublicKeys(ctx)
	if err != nil {
		return nil, err
	}

	if p, ok := keys[keyHash]; ok {
		return p, nil
	}
	return nil, ErrVaultNotFound
}

// GetPublicKey retrieve the public key from a vault
func (s *Signatory) GetPublicKey(ctx context.Context, keyHash string) (*PublicKey, error) {
	p, err := s.getPublicKey(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	enc, err := tezos.EncodePublicKey(p.key.PublicKey())
	if err != nil {
		return nil, err
	}

	return &PublicKey{
		PublicKey:     enc,
		PublicKeyHash: keyHash,
		VaultName:     p.vault.Name(),
		ID:            p.key.ID(),
		Policy:        s.fetchPolicyOrDefault(keyHash),
	}, nil
}

// Unlock unlock all the vaults
func (s *Signatory) Unlock(ctx context.Context) error {
	for _, v := range s.vaults {
		if unlocker, ok := v.(vault.Unlocker); ok {
			if err := unlocker.Unlock(ctx); err != nil {
				return err
			}
		}
	}
	s.cache.drop()
	return nil
}

// Config represents Signatory configuration
type Config struct {
	Policy      map[string]*Policy
	Vaults      map[string]*config.VaultConfig
	Interceptor SignInterceptor
	Watermark   Watermark
	Logger      log.FieldLogger
}

// NewSignatory returns Signatory instance
func NewSignatory(ctx context.Context, c *Config) (*Signatory, error) {
	s := &Signatory{
		config: *c,
		vaults: make(map[string]vault.Vault, len(c.Vaults)),
	}

	// Initialize vaults
	for name, vc := range c.Vaults {
		l := s.logger().WithFields(log.Fields{
			logVault:     vc.Driver,
			logVaultName: name,
		})

		l.Info("Initializing vault")

		v, err := vault.NewVault(ctx, vc.Driver, &vc.Config)
		if err != nil {
			return nil, err
		}
		s.vaults[name] = v
	}

	return s, nil
}

// Ready returns true if all backends are ready
func (s *Signatory) Ready(ctx context.Context) (bool, error) {
	for _, v := range s.vaults {
		if rc, ok := v.(vault.ReadinessChecker); ok {
			if ok, err := rc.Ready(ctx); !ok || err != nil {
				return ok, err
			}
		}
	}
	return true, nil
}

// PreparePolicy prepares policy data by hashing keys etc
func PreparePolicy(src config.TezosConfig) (map[string]*Policy, error) {
	policy := make(map[string]*Policy, len(src))
	for k, v := range src {
		pol := Policy{
			LogPayloads: v.LogPayloads,
		}

		if v.AllowedKinds != nil {
			pol.AllowedKinds = make([]string, len(v.AllowedKinds))
			copy(pol.AllowedKinds, v.AllowedKinds)
			sort.Strings(pol.AllowedKinds)
		}

		if v.AllowedOperations != nil {
			pol.AllowedOperations = make([]string, len(v.AllowedOperations))
			copy(pol.AllowedOperations, v.AllowedOperations)
			sort.Strings(pol.AllowedOperations)
		}

		if v.AuthorizedKeys != nil {
			keys := v.AuthorizedKeys.List()
			pol.AuthorizedKeys = make([]string, len(keys))
			for i, k := range keys {
				pub, err := tezos.ParsePublicKey(k)
				if err != nil {
					return nil, err
				}
				pol.AuthorizedKeys[i], err = tezos.EncodePublicKeyHash(pub)
				if err != nil {
					return nil, err
				}
			}
		}
		policy[k] = &pol
	}
	return policy, nil
}
