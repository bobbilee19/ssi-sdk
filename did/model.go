package did

import (
	"fmt"
	"reflect"

	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/cryptosuite"

	"github.com/TBD54566975/ssi-sdk/util"
)

const (
	KnownDIDContext string = "https://www.w3.org/ns/did/v1"

	// Base58BTCMultiBase Base58BTC https://github.com/multiformats/go-multibase/blob/master/multibase.go
	Base58BTCMultiBase = multibase.Base58BTC

	// Multicodec reference https://github.com/multiformats/multicodec/blob/master/table.csv

	Ed25519MultiCodec   = multicodec.Ed25519Pub
	X25519MultiCodec    = multicodec.X25519Pub
	Secp256k1MultiCodec = multicodec.Secp256k1Pub
	P256MultiCodec      = multicodec.P256Pub
	P384MultiCodec      = multicodec.P384Pub
	P521MultiCodec      = multicodec.P521Pub
	RSAMultiCodec       = multicodec.RsaPub
	SHA256MultiCodec    = multicodec.Sha2_256
)

// DIDResolutionResult encapsulates the tuple of a DID resolution https://www.w3.org/TR/did-core/#did-resolution
type DIDResolutionResult struct {
	DIDResolutionMetadata
	DIDDocument
	DIDDocumentMetadata
}

// DIDDocumentMetadata https://www.w3.org/TR/did-core/#did-document-metadata
type DIDDocumentMetadata struct {
	Created       string `json:"created,omitempty" validate:"datetime"`
	Updated       string `json:"updated,omitempty" validate:"datetime"`
	Deactivated   bool   `json:"deactivated,omitempty"`
	NextUpdate    string `json:"nextUpdate,omitempty"`
	VersionID     string `json:"versionId,omitempty"`
	NextVersionID string `json:"nextVersionId,omitempty"`
	EquivalentID  string `json:"equivalentId,omitempty"`
	CanonicalID   string `json:"canonicalId,omitempty"`
}

func (s *DIDDocumentMetadata) IsValid() bool {
	return util.NewValidator().Struct(s) == nil
}

// ResolutionError https://www.w3.org/TR/did-core/#did-resolution-metadata
type ResolutionError struct {
	Code                       string `json:"code"`
	InvalidDID                 bool   `json:"invalidDid"`
	NotFound                   bool   `json:"notFound"`
	RepresentationNotSupported bool   `json:"representationNotSupported"`
}

// DIDResolutionMetadata https://www.w3.org/TR/did-core/#did-resolution-metadata
type DIDResolutionMetadata struct {
	ContentType string
	Error       *ResolutionError
}

// DIDDocument is a representation of the did core specification https://www.w3.org/TR/did-core
// TODO(gabe) enforce validation of DID syntax https://www.w3.org/TR/did-core/#did-syntax
type DIDDocument struct {
	Context any `json:"@context,omitempty"`
	// As per https://www.w3.org/TR/did-core/#did-subject intermediate representations of DID Documents do not
	// require an ID property. The provided test vectors demonstrate IRs. As such, the property is optional.
	ID                   string                  `json:"id,omitempty"`
	Controller           string                  `json:"controller,omitempty"`
	AlsoKnownAs          string                  `json:"alsoKnownAs,omitempty"`
	VerificationMethod   []VerificationMethod    `json:"verificationMethod,omitempty" validate:"dive"`
	Authentication       []VerificationMethodSet `json:"authentication,omitempty" validate:"dive"`
	AssertionMethod      []VerificationMethodSet `json:"assertionMethod,omitempty" validate:"dive"`
	KeyAgreement         []VerificationMethodSet `json:"keyAgreement,omitempty" validate:"dive"`
	CapabilityInvocation []VerificationMethodSet `json:"capabilityInvocation,omitempty" validate:"dive"`
	CapabilityDelegation []VerificationMethodSet `json:"capabilityDelegation,omitempty" validate:"dive"`
	Services             []Service               `json:"service,omitempty" validate:"dive"`
}

type VerificationMethod struct {
	ID              string                `json:"id" validate:"required"`
	Type            cryptosuite.LDKeyType `json:"type" validate:"required"`
	Controller      string                `json:"controller" validate:"required"`
	PublicKeyBase58 string                `json:"publicKeyBase58,omitempty"`
	// must conform to https://datatracker.ietf.org/doc/html/rfc7517
	PublicKeyJWK *crypto.PublicKeyJWK `json:"publicKeyJwk,omitempty" validate:"omitempty,dive"`
	// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03
	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`
	// for PKH DIDs - https://github.com/w3c-ccg/did-pkh/blob/90b28ad3c18d63822a8aab3c752302aa64fc9382/did-pkh-method-draft.md
	BlockchainAccountID string `json:"blockchainAccountId,omitempty"`
}

// VerificationMethodSet is a union type supporting the `authentication`, `assertionMethod`, `keyAgreement`,
// `capabilityInvocation`, and `capabilityDelegation` types.
// A set of one or more verification methods. Each verification method MAY be embedded or referenced.
// TODO(gabe) consider changing this to a custom unmarshaler https://stackoverflow.com/a/28016508
type VerificationMethodSet any

// Service is a property compliant with the did-core spec https://www.w3.org/TR/did-core/#services
type Service struct {
	ID   string `json:"id" validate:"required"`
	Type string `json:"type" validate:"required"`
	// A string, map, or set composed of one or more strings and/or maps
	// All string values must be valid URIs
	ServiceEndpoint any      `json:"serviceEndpoint" validate:"required"`
	RoutingKeys     []string `json:"routingKeys,omitempty"`
	Accept          []string `json:"accept,omitempty"`
}

func (s *Service) IsValid() bool {
	return util.NewValidator().Struct(s) == nil
}

func (d *DIDDocument) IsEmpty() bool {
	if d == nil {
		return true
	}
	return reflect.DeepEqual(d, &DIDDocument{})
}

func (d *DIDDocument) IsValid() error {
	return util.NewValidator().Struct(d)
}

// KeyTypeToLDKeyType converts crypto.KeyType to cryptosuite.LDKeyType
func KeyTypeToLDKeyType(kt crypto.KeyType) (cryptosuite.LDKeyType, error) {
	switch kt {
	case crypto.Ed25519:
		return cryptosuite.Ed25519VerificationKey2018, nil
	case crypto.X25519:
		return cryptosuite.X25519KeyAgreementKey2019, nil
	case crypto.SECP256k1:
		return cryptosuite.ECDSASECP256k1VerificationKey2019, nil
	case crypto.P256, crypto.P384, crypto.P521, crypto.RSA:
		return cryptosuite.JSONWebKey2020Type, nil
	default:
		return "", fmt.Errorf("keyType %+v failed to convert to LDKeyType", kt)
	}
}
