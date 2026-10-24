package evidence

import "github.com/saphan-ai/proven/internal/envdoc"

// Environment lists the evidence-pack settings in the order the page shows.
// Each Name is the Env constant beside this function. Both hold a file path,
// so neither is a secret value: the key itself is read from the file and is
// never printed, logged or written into a pack.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvPackKey,
			Default:     "none (required with `--evidence-pack`)",
			Description: "path to the PEM file of the organisation's Ed25519 private key (PKCS #8) that signs a pack. A missing or unreadable file, or one that group or others can reach (fix: `chmod 600`), is named and the review does not start",
		},
		{
			Name:        EnvPackPublicKey,
			Default:     "none (required by `proven verify` unless `--public-key` is given)",
			Description: "path to the PEM file of the Ed25519 public key (PKIX) that verifies a pack",
		},
	}
}
