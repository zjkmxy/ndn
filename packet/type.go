package packet

const (
	// common
	NAME uint64 = iota
	// interest
	INTEREST
	SELECTOR
	NONCE
	MIN_PREFIX_COMPONENT
	MAX_PREFIX_COMPONENT
	PUBLISHER_PUBLICKEY
	EXCLUDE
	ANY
	CHILD_SELECTOR
	MUST_BE_FRESH
	SCOPE
	INTEREST_LIFETIME
	//data
	DATA
	META_INFO
	CONTENT_TYPE
	FRESHNESS_PERIOD
	CONTENT
	SIGNATURE
	DIGEST_SHA256
	SIGNATURE_SHA256_WITH_RSA
	SIGNATURE_SHA256_WITH_RSA_AND_MERKLE
	KEY_LOCATOR
	CERTIFICATE_NAME
)
