(window.webpackJsonp=window.webpackJsonp||[]).push([[7],{142:function(e,n,t){"use strict";t.r(n),t.d(n,"frontMatter",(function(){return r})),t.d(n,"metadata",(function(){return s})),t.d(n,"rightToc",(function(){return l})),t.d(n,"default",(function(){return d}));var a=t(2),i=t(9),o=(t(0),t(153)),r={},s={id:"README",title:"README",description:"Signatory",source:"@site/../docs/README.md",permalink:"/docs/README",editUrl:"https://github.com/ecadlabs/signatory/edit/master/website/../docs/README.md"},l=[{value:"Configuration",id:"configuration",children:[{value:"Configuration example",id:"configuration-example",children:[]}]},{value:"Backends",id:"backends",children:[]},{value:"Signatory service",id:"signatory-service",children:[{value:"Prometheus metrics and health service",id:"prometheus-metrics-and-health-service",children:[]},{value:"Testing",id:"testing",children:[]}]},{value:"Signatory command line tool",id:"signatory-command-line-tool",children:[{value:"Import a private key",id:"import-a-private-key",children:[]},{value:"List keys",id:"list-keys",children:[]}]}],c={rightToc:l};function d(e){var n=e.components,t=Object(i.a)(e,["components"]);return Object(o.b)("wrapper",Object(a.a)({},c,t,{components:n,mdxType:"MDXLayout"}),Object(o.b)("h1",{id:"signatory"},"Signatory"),Object(o.b)("h2",{id:"configuration"},"Configuration"),Object(o.b)("p",null,"Signatory configuration is specified in a YAML file. Use the ",Object(o.b)("inlineCode",{parentName:"p"},"signatory.yaml")," file as a template to getting started."),Object(o.b)("p",null,"Each backend can be configured with one of more instances of the backend. The operator can add as many new backends as they wish, just append to the list."),Object(o.b)("p",null,"Configuration file is shared between ",Object(o.b)("inlineCode",{parentName:"p"},"signatory")," and ",Object(o.b)("inlineCode",{parentName:"p"},"signatory-cli"),"."),Object(o.b)("h3",{id:"configuration-example"},"Configuration example"),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{className:"language-yaml"}),'server:\n  # Address for the main HTTP server to listen on\n  address: :6732\n  # Address for the utility HTTP server to listen on\n  utility_address: :9583\n\nvaults:\n  # Name is used to identify backend during import process\n  kms:\n    driver: cloudkms\n    config:\n      # See backend specific documentation\n      project: signatory\n      location: europe-north1\n      key_ring: hsm-ring\n  azure:\n    driver: azure\n    config:\n      # See backend specific documentation\n      vault: https://signatory.vault.azure.net/\n      tenant_id: cf5dd0ba-d3a3-4f3f-a688-06d12672f8ed\n      client_id: 5d29a974-edd0-4659-b933-7d9c56726649\n      client_pkcs12_certificate: principal.pfx\n  yubi:\n    driver: yubihsm\n    config:\n      # See backend specific documentation\n      address: localhost:12345\n      password: password\n      auth_key_id: 1\n\n# List enabled public keys hashes here\ntezos:\n  # Default policy allows "block" and "endorsement" operations\n  tz1Wz4ZabKRsz842Xuzy4a7CcWADfPVsPKus:\n  # Explicit policy\n  tz3MhmeqpudUqEX8PYTbNDF3CVcnnjNQoo8N:\n    # Setting `log_payloads` to `true` will cause Signatory to log operation\n    # payloads to `stdout`. This may be desirable for audit and investigative\n    # purposes.\n    log_payloads: true\n    allowed_operations:\n      # List of [generic, block, endorsement]\n      - generic\n      - block\n      - endorsement\n    allowed_kinds:\n      # List of [endorsement, ballot, reveal, transaction, origination, delegation, seed_nonce_revelation, activate_account]\n      - transaction\n      - endorsement\n')),Object(o.b)("h2",{id:"backends"},"Backends"),Object(o.b)("ul",null,Object(o.b)("li",{parentName:"ul"},Object(o.b)("a",Object(a.a)({parentName:"li"},{href:"/pkg/vault/cloudkms/README.md"}),"Google Cloud KMS")),Object(o.b)("li",{parentName:"ul"},Object(o.b)("a",Object(a.a)({parentName:"li"},{href:"/pkg/vault/azure/README.md"}),"Azure")),Object(o.b)("li",{parentName:"ul"},Object(o.b)("a",Object(a.a)({parentName:"li"},{href:"/pkg/vault/yubi/README.md"}),"YubiHSM2"))),Object(o.b)("hr",null),Object(o.b)("h2",{id:"signatory-service"},"Signatory service"),Object(o.b)("p",null,"Signatory service is used for signing operations and implements Tezos specific HTTP external signer API"),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{}),'A Tezos Remote Signer for signing block-chain operations with private keys\n\nUsage:\n  signatory [flags]\n  signatory [command]\n\nAvailable Commands:\n  help        Help about any command\n  serve       Run a server\n\nFlags:\n  -c, --config string   Config file path (default "signatory.yaml")\n  -h, --help            help for signatory\n      --log string      Log level: [error, warn, info, debug, trace] (default "info")\n')),Object(o.b)("h3",{id:"prometheus-metrics-and-health-service"},"Prometheus metrics and health service"),Object(o.b)("p",null,"Signatory exposes Prometheus metrics and a health status on address specified in ",Object(o.b)("inlineCode",{parentName:"p"},"utility_address")," configuration parameter. Default value is ",Object(o.b)("inlineCode",{parentName:"p"},":9583"),"."),Object(o.b)("h4",{id:"prometheus-metrics"},"Prometheus metrics"),Object(o.b)("p",null,"Metrics include counters and histograms that track signing operations and errors."),Object(o.b)("p",null,"The metrics are intended to be scraped using the Prometheus time series database. We also publish a ready-made Grafana dashboard which users can use to visualize the operation of their signing operations. (TODO: publish Grafana dashboard)"),Object(o.b)("p",null,Object(o.b)("inlineCode",{parentName:"p"},"localhost:9583/metrics")),Object(o.b)("h4",{id:"health-service"},"Health service"),Object(o.b)("p",null,"Health service endpoint can be used to test if the service is running correctly, and ready to sign requests."),Object(o.b)("p",null,"This endpoints is useful for use in monitoring, or declarative tests as part of deployment playbooks or kubernetes manifests."),Object(o.b)("p",null,Object(o.b)("inlineCode",{parentName:"p"},"localhost:9583/healthz")),Object(o.b)("h3",{id:"testing"},"Testing"),Object(o.b)("p",null,"To test the signing operation, you can send a post to signatory. In this example, we are sending a dummy operation of type ",Object(o.b)("inlineCode",{parentName:"p"},"02"),", which is a ",Object(o.b)("inlineCode",{parentName:"p"},"endorsement")," operation type. "),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{className:"language-sh"}),"curl -XPOST \\\n    -d '\"02111111111111111111\"' \\\n    localhost:8003/keys/tz3Tm6UTWmPAZJaNSPAQNiMiyFSHnRXrkcHj\n")),Object(o.b)("p",null,"If you receive an error from curl and on the signatory console, you will have to investigate. If it was successful, you should see output similar to:"),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{className:"language-json"}),'{"signature":"p2sigR4JTRTMkT4XC4NgVuGdhZDbgaaSZpNPUserkyMCTY1GQJTFpCuihFRVk9n7YaNjA5U3cNcvJPRm7C9G5A1hsLsesVPcMu"}\n')),Object(o.b)("hr",null),Object(o.b)("h2",{id:"signatory-command-line-tool"},"Signatory command line tool"),Object(o.b)("p",null,"Signatory service is used for importing of private keys and obtaining information about available key pairs and their policies."),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{}),'A Tezos Remote Signer for signing block-chain operations with private keys\n\nUsage:\n  signatory-cli import [flags]\n\nFlags:\n  -h, --help              help for import\n  -o, --opt string        Options to be passed to the backend. Syntax: key:val[,...]\n      --password string   Password for private key(s)\n      --vault string      Vault name for importing\n\nGlobal Flags:\n  -c, --config string   Config file path (default "signatory.yaml")\n      --log string      Log level: [error, warn, info, debug, trace] (default "info")\n')),Object(o.b)("h3",{id:"import-a-private-key"},"Import a private key"),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{className:"language-sh"}),"signatory-cli -c CONFIG import --vault VAULT PRIVATE_KEY\n")),Object(o.b)("p",null,"Example:"),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{className:"language-sh"}),"signatory-cli -c signatory.yaml import --vault yubi edsk3rsARzj7f8PEHXXUbLigMDCww75nPnzbFmSz19TLwzrYzF8uCB\n")),Object(o.b)("h3",{id:"list-keys"},"List keys"),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{className:"language-sh"}),"signatory-cli -c CONFIG list\n")),Object(o.b)("p",null,"Example:"),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{className:"language-sh"}),"signatory-cli -c signatory.yaml list\n")),Object(o.b)("p",null,"Example output:"),Object(o.b)("pre",null,Object(o.b)("code",Object(a.a)({parentName:"pre"},{}),"INFO[0000] Initializing vault                            vault=cloudkms vault_name=kms\nINFO[0000] Initializing vault                            vault=azure vault_name=azure\nPublic Key Hash:    tz3VfoCwiQyMNYnaseFLFAjN9AQJQnhvddjG\nVault:              CloudKMS\nID:                 projects/signatory-testing/locations/europe-north1/keyRings/hsm-ring/cryptoKeys/hsm-key/cryptoKeyVersions/1\nAllowed Operations: [block endorsement]\nAllowed Kinds:      []\n\nPublic Key Hash:    tz3ZqyLdKy2doLbw7yghLPz2TWWZdxeLGKVx\nVault:              CloudKMS\nID:                 projects/signatory-testing/locations/europe-north1/keyRings/hsm-ring/cryptoKeys/hsm-key/cryptoKeyVersions/2\n*DISABLED*\n\nPublic Key Hash:    tz3aTwpna6m9qsw4YZVFad1nsm5cGgWHVQ8R\nVault:              CloudKMS\nID:                 projects/signatory-testing/locations/europe-north1/keyRings/hsm-ring/cryptoKeys/signatory-imported-1RG8mJUH8P5ncMEMypfkno98Gpq/cryptoKeyVersions/1\nAllowed Operations: [block endorsement generic]\nAllowed Kinds:      [endorsement transaction]\n\nPublic Key Hash:    tz3VkMSRVjLwEoUgZNJwjoD6YHeBDXyWiBaY\nVault:              Azure\nID:                 https://signatory.vault.azure.net/keys/key0/fa9607734e58485181d19da901e725b9\n*DISABLED*\n\n")))}d.isMDXComponent=!0}}]);