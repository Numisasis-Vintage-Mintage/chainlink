package vrfkey

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"

	"chainlink/core/services/signatures/secp256k1"
	"chainlink/core/utils"
)

// PublicKey is a secp256k1 point in compressed format
type PublicKey struct {
	comprPoint [CompressedPublicKeyLength]byte
}

// CompressedPublicKeyLength is the length of a secp256k1 public key's x
// ordinate as a uint256, concatenated with 00 if y is even, 01 if odd.
const CompressedPublicKeyLength = 33

func init() {
	if CompressedPublicKeyLength != (&secp256k1.Secp256k1{}).Point().MarshalSize() {
		panic("disparity in expected public key lengths")
	}
}

// Set sets k to the public key represented by l
func (k *PublicKey) Set(l PublicKey) {
	if copy(k.comprPoint[:], l.comprPoint[:]) != CompressedPublicKeyLength {
		panic(fmt.Errorf("failed to copy entire public key %x to %x", l, k))
	}
}

// CompressedPoint returns the compressed secp256k1 point corresponding to k
func (k *PublicKey) SerializeCompressed() []byte {
	return k.comprPoint[:]
}

// Point returns the secp256k1 point corresponding to k
func (k *PublicKey) Point() (kyber.Point, error) {
	p := (&secp256k1.Secp256k1{}).Point()
	return p, p.UnmarshalBinary(k.comprPoint[:])
}

// NewPublicKey returns the PublicKey corresponding to rawKey
func NewPublicKey(rawKey [CompressedPublicKeyLength]byte) (*PublicKey, error) {
	rv := PublicKey{comprPoint: rawKey}
	_, err := rv.Point()
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid rawKey given to NewPublicKey.")
	}
	return &rv, nil
}

// NewPublicKeyFromSlice returns the PublicKey corresponding to rawKey or errors
func NewPublicKeyFromSlice(rawKey []byte) (*PublicKey, error) {
	if l := len(rawKey); l != CompressedPublicKeyLength {
		return nil, fmt.Errorf("wrong length for public key: %s of length %d", rawKey, l)
	}
	var rawKeyArray [CompressedPublicKeyLength]byte
	if c := copy(rawKeyArray[:], rawKey[:]); c != CompressedPublicKeyLength {
		panic(fmt.Errorf("failed to copy entire key to return value"))
	}
	return NewPublicKey(rawKeyArray)
}

// NewPublicKeyFromHex returns the PublicKey encoded by 0x-hex string hex, or errors
func NewPublicKeyFromHex(hex string) (*PublicKey, error) {
	rawKey, err := hexutil.Decode(hex)
	if err != nil {
		return nil, err
	}
	return NewPublicKeyFromSlice(rawKey)
}

// SetFromHex sets k to the public key represented by hex, which must represent
// the uncompressed binary format
func (k *PublicKey) SetFromHex(hex string) error {
	nk, err := NewPublicKeyFromHex(hex)
	if err != nil {
		return err
	}
	k.Set(*nk)
	return nil
}

// String returns k's binary uncompressed representation, as 0x-hex
func (k PublicKey) String() string {
	return hexutil.Encode(k.comprPoint[:])
}

// Hash returns the solidity Keccak256 hash of k. Corresponds to hashOfKey on
// VRFCoordinator.
func (k *PublicKey) Hash() common.Hash {
	return utils.MustHash(string(k.comprPoint[:]))
}

// Address returns the Ethereum address of k
func (k *PublicKey) Address() common.Address {
	return common.BytesToAddress(k.Hash().Bytes()[12:])
}
