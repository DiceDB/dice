package eval

import (
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
)

type Capacity uint64
type BucketSize uint8
type MaxIterations uint16
type Expansion uint16

const (
	defaultBucketSize    BucketSize    = 2
	defaultMaxIterations MaxIterations = 20
	defaultExpansion     Expansion     = 1
)

var (
	errCFInvalidCapacityType      = diceerrors.NewErr("only integer values can be provided for capacity")
	errCFInvalidCapacity          = diceerrors.NewErr("invalid capacity value provided")
	errCFInvalidBucketSize        = diceerrors.NewErr("invalid bucket size provided")
	errCFInvalidBucketVal         = diceerrors.NewErr("only integer values can be provided for bucket size")
	errCFInvalidMaxIterationsType = diceerrors.NewErr("only integer values can be provided for MAXITERATIONS")
	errCFInvalidMaxIterationsVal  = diceerrors.NewErr("invalid MAXITERATIONS value provided")
	errCFInvalidExpansionType     = diceerrors.NewErr("only integer values can be provided for expansion")
	errCFInvalidExpansionVal      = diceerrors.NewErr("invalid EXPANSION value provided")
	errCFInvalidOption            = diceerrors.NewErr("invalid option provided")
)

type CuckooOpts struct {
	capacity      Capacity
	bucketSize    BucketSize
	maxIterations MaxIterations
	expansion     Expansion
}

type CuckooFilter struct {
	opts            *CuckooOpts
	buckets         []bucket
	bucketIndexMask uint
	count           uint
}

// Bucket operations implement starts

type fingerPrint uint8
type bucket []fingerPrint

const (
	// bucketSize = 4
	fingerPrintSizeBits = 16
	maxFingerPrint      = (1 << fingerPrintSizeBits) - 1
	nullFingerPrint     = 0
	maxDisplacements    = 500
)

func newCuckooOpts(args []string, useDefaults bool) (*CuckooOpts, error) {

	capacity, err := strconv.ParseUint(args[1], 10, 64)

	if err != nil {
		return nil, errCFInvalidCapacityType
	}

	if capacity < 1 {
		return nil, errCFInvalidCapacity
	}
	opts := &CuckooOpts{
		capacity:      Capacity(capacity),
		bucketSize:    defaultBucketSize,
		maxIterations: defaultMaxIterations,
		expansion:     defaultExpansion,
	}
	if !useDefaults {
		for i := 2; i < len(args); i++ {
			arg := strings.ToUpper(args[i])
			switch arg {
			case "BUCKETSIZE":
				if i+1 < len(args) {
					bucketSize, err := strconv.ParseUint(args[i+1], 10, 8)
					// @TODO err != nil
					if err != nil {
						return nil, errCFInvalidBucketVal
					}
					// bucketSize > 0
					if bucketSize < 1 || bucketSize > 255 {
						return nil, errCFInvalidBucketSize
					}
					opts.bucketSize = BucketSize(bucketSize)
					i++
				} else {
					return nil, diceerrors.NewErr("missing value for BUCKETSIZE")
				}
			case "MAXITERATIONS":
				if i+1 < len(args) {
					maxIterations, err := strconv.ParseUint(args[i+1], 10, 16)
					if err != nil {
						return nil, errCFInvalidMaxIterationsType
					}

					if maxIterations < 1 || maxIterations > 65535 {
						return nil, errCFInvalidMaxIterationsVal
					}
					opts.maxIterations = MaxIterations(maxIterations)
					i++
				} else {
					return nil, diceerrors.NewErr("missing value for MAXITERATIONS")
				}
			case "EXPANSION":
				if i+1 < len(args) {
					expansion, err := strconv.ParseUint(args[i+1], 10, 16)
					if err != nil {
						return nil, errCFInvalidExpansionType
					}

					if expansion < 0 || expansion > 32768 {
						return nil, errCFInvalidExpansionVal
					}
					opts.expansion = Expansion(expansion)
					i++
				} else {
					return nil, diceerrors.NewErr("missing value for EXPANSION")
				}
			default:
				return nil, errCFInvalidOption
			}

		}
	}

	return opts, nil
}

// @TODO also return , error
func newCuckooFilter(opts *CuckooOpts) (*CuckooFilter, error) {
	numOfBuckets := uint64(opts.capacity) / uint64(opts.bucketSize)
	numOfBuckets = getNextPow2(numOfBuckets)

	if numOfBuckets == 0 {
		numOfBuckets = 1
	}

	if float64(opts.capacity)/float64(numOfBuckets*uint64(opts.bucketSize)) > 0.95 {
		numOfBuckets <<= 1
	}

	buckets := make([]bucket, numOfBuckets)

	for i := range buckets {
		buckets[i] = make(bucket, opts.bucketSize)
	}

	return &CuckooFilter{
		buckets:         buckets,
		bucketIndexMask: uint(len(buckets) - 1),
		opts:            opts,
	}, nil

}

// command evaluators

func evalCFRESERVE(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("CF.RESERVE")
	}

	useDefaults := false

	if len(args) == 2 {
		useDefaults = true
	}

	opts, err := newCuckooOpts(args, useDefaults)
	// InitCF
	if err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CF.RESERVE' command", err)
	}

	cf, _ := newCuckooFilter(opts)
	obj := store.NewObj(cf, -1, object.ObjTypeBitSet, object.ObjEncodingCF)

	store.Put(args[0], obj)
	return clientio.RespOK
}

// @TODO init filter if does not exist here
func evalCFADD(args []string, store *dstore.Store) []byte {

	if len(args) != 2 {
		return diceerrors.NewErrArity("CF.ADD")
	}

	key := args[0]
	item := []byte(args[1])

	cfInstance := store.Get(key)

	cf, ok := cfInstance.Value.(*CuckooFilter)

	if !ok {
		return clientio.RespEmptyArray
	}

	if added := cf.add(item); !added {
		return clientio.RespEmptyArray
	}

	return clientio.RespOne
}

func evalCFEXISTS(args []string, store *dstore.Store) []byte {
	if len(args) != 2 {
		return diceerrors.NewErrArity("CF.EXISTS")
	}
	key := args[0]
	item := []byte(args[1])
	cfInstance := store.Get(key)
	cf, ok := cfInstance.Value.(*CuckooFilter)
	if !ok {
		return clientio.RespEmptyArray
	}

	if exists := cf.lookup(item); !exists {
		return clientio.RespZero
	}

	return clientio.RespOne
}
