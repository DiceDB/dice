package types

import (
	"math"
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dicedb-go/wire"
	"github.com/emirpasic/gods/queues/priorityqueue"
	"github.com/mmcloughlin/geohash"
)

// Bit precision - Same as redis (https://github.com/redis/redis/blob/5d0d64b062c160093dc287ed5d18ec7c807873cf/src/geohash_helper.c#L213)
const BIT_PRECISION = 52

// Earth's radius in meters
const EARTH_RADIUS float64 = 6372797.560856

// The maximum/minumum projected coordinate value (in meters) in the Web Mercator projection (EPSG:3857)
// Earth’s equator: ~40,075 km → half of that = ~20,037 km
// The Mercator projection transforms the globe into a square map.
// MERCATOR_MAX is the extent of that square in meters.
const MERCATOR_MAX float64 = 20037726.37
const MERCATOR_MIN float64 = -20037726.37

// Limits from EPSG:900913 / EPSG:3785 / OSGEO:41001
const LAT_MIN float64 = -85.05112878
const LAT_MAX float64 = 85.05112878
const LONG_MIN float64 = -180
const LONG_MAX float64 = 180

// GEO Neighbors
type Neighbors struct {
	North     uint64
	NorthEast uint64
	East      uint64
	SouthEast uint64
	South     uint64
	SouthWest uint64
	West      uint64
	NorthWest uint64
	Center    uint64
	Precision uint
}

func CreateNeighborsFromArray(arr []uint64) *Neighbors {
	neightbors := Neighbors{}
	if len(arr) < 8 {
		return &neightbors
	}

	neightbors.North = arr[0]
	neightbors.NorthEast = arr[1]
	neightbors.East = arr[2]
	neightbors.SouthEast = arr[3]
	neightbors.South = arr[4]
	neightbors.SouthWest = arr[5]
	neightbors.West = arr[6]
	neightbors.NorthWest = arr[7]
	neightbors.Center = arr[8]

	return &neightbors
}

func (neightbors *Neighbors) ToArray() [9]uint64 {
	arr := [9]uint64{}

	arr[0] = neightbors.North
	arr[1] = neightbors.NorthEast
	arr[2] = neightbors.East
	arr[3] = neightbors.SouthEast
	arr[4] = neightbors.South
	arr[5] = neightbors.SouthWest
	arr[6] = neightbors.West
	arr[7] = neightbors.NorthWest
	arr[8] = neightbors.Center

	return arr
}

// Struct to save coordinates
type GeoCoordinate struct {
	Hash      uint64
	Longitude float64
	Latitude  float64
}

func NewGeoCoordinateFromLonLat(longitude, latitude float64) (*GeoCoordinate, error) {
	if err := ValidateLonLat(longitude, latitude); err != nil {
		return nil, err
	}
	hash := EncodeHash(longitude, latitude)
	return &GeoCoordinate{
		Hash:      hash,
		Longitude: longitude,
		Latitude:  latitude,
	}, nil
}

func NewGeoCoordinateFromHash(hash uint64) (*GeoCoordinate, error) {
	longitude, latitude := DecodeHash(hash)
	if err := ValidateLonLat(longitude, latitude); err != nil {
		return nil, err
	}
	return &GeoCoordinate{
		Hash:      hash,
		Longitude: longitude,
		Latitude:  latitude,
	}, nil
}

func (coord *GeoCoordinate) GetLatDistanceFromCoordinate(otherCoord *GeoCoordinate) float64 {
	return EARTH_RADIUS * math.Abs(DegToRad(otherCoord.Latitude)-DegToRad(coord.Latitude))
}

func (coord *GeoCoordinate) GetDistanceFromCoordinate(otherCoord *GeoCoordinate) float64 {
	lon1r := DegToRad(coord.Longitude)
	lat1r := DegToRad(coord.Latitude)

	lon2r := DegToRad(otherCoord.Longitude)
	lat2r := DegToRad(otherCoord.Latitude)

	v := math.Sin((lon2r - lon1r) / 2)
	// if v == 0 we can avoid doing expensive math when lons are practically the same
	if v == 0.0 {
		return coord.GetLatDistanceFromCoordinate(otherCoord)
	}

	u := math.Sin((lat2r - lat1r) / 2)

	a := u*u + math.Cos(lat1r)*math.Cos(lat2r)*v*v

	return 2.0 * EARTH_RADIUS * math.Asin(math.Sqrt(a))
}

// This saves only the Hash of Longitude and Latitude
type GeoRegistry struct {
	*SortedSet
}

func NewGeoRegistry() *GeoRegistry {
	return &GeoRegistry{
		SortedSet: NewSortedSet(),
	}
}

func (geoReg *GeoRegistry) Add(coordinates []*GeoCoordinate, members []string, params map[Param]string) (int64, error) {

	hashArr := []int64{}
	for _, coord := range coordinates {
		hashArr = append(hashArr, int64(coord.Hash))
	}

	// Note: Validation of the params is done in the SortedSet.ZADD method
	return geoReg.ZADD(hashArr, members, params)

}

func (geoReg *GeoRegistry) GetDistanceBetweenMembers(member1 string, member2 string, unit Param) (float64, error) {
	node1 := geoReg.GetByKey(member1)
	node2 := geoReg.GetByKey(member2)

	if node1 == nil || node2 == nil {
		return 0, nil
	}

	hash1 := node1.Score()
	hash2 := node2.Score()

	coord1, _ := NewGeoCoordinateFromHash(uint64(hash1))
	coord2, _ := NewGeoCoordinateFromHash(uint64(hash2))

	dist, err := ConvertDistance(coord1.GetDistanceFromCoordinate(coord2), unit)

	if err != nil {
		return 0, err
	}

	return dist, nil

}

// This returns all the nodes which are in the given shape
func (geoReg *GeoRegistry) SearchElementsWithinShape(params map[Param]string, nonParams []string) ([]*wire.GEOElement, error) {
	unit := GetUnitTypeFromParsedParams(params)
	if len(unit) == 0 {
		return nil, errors.ErrInvalidUnit(string(unit))
	}

	// Return error if both FROMLONLAT & FROMMEMBER are set
	if params[FROMLONLAT] != "" && params[FROMMEMBER] != "" {
		return nil, errors.ErrInvalidSetOfOptions(string(FROMLONLAT), string(FROMMEMBER))
	}

	// Return error if none of FROMLONLAT & FROMMEMBER are set
	if params[FROMLONLAT] == "" && params[FROMMEMBER] == "" {
		return nil, errors.ErrNeedOneOfTheOptions(string(FROMLONLAT), string(FROMMEMBER))
	}

	// Return error if both BYBOX & BYRADIUS are set
	if params[BYBOX] != "" && params[BYRADIUS] != "" {
		return nil, errors.ErrInvalidSetOfOptions(string(BYBOX), string(BYRADIUS))
	}

	// Return error if none of BYBOX & BYRADIUS are set
	if params[BYBOX] == "" && params[BYRADIUS] == "" {
		return nil, errors.ErrNeedOneOfTheOptions(string(BYBOX), string(BYRADIUS))
	}

	// Return error if ANY is used without COUNT
	if params[ANY] != "" && params[COUNT] == "" {
		return nil, errors.ErrGeneral("ANY argument requires COUNT argument")
	}

	// Return error if Both ASC & DESC are used
	if params[ASC] != "" && params[DESC] != "" {
		return nil, errors.ErrGeneral("Use one of ASC or DESC")
	}

	// Fetch Longitute and Latitude based on FROMLONLAT & FROMMEMBER param
	var centerCoordinate *GeoCoordinate
	var err error

	// Fetch Longitute and Latitude from params
	if params[FROMLONLAT] != "" {
		if len(nonParams) < 2 {
			return nil, errors.ErrWrongArgumentCount("GEOSEARCH")
		}
		lon, errLon := strconv.ParseFloat(nonParams[0], 10)
		lat, errLat := strconv.ParseFloat(nonParams[1], 10)

		if errLon != nil || errLat != nil {
			return nil, errors.ErrInvalidNumberFormat
		}

		centerCoordinate, err = NewGeoCoordinateFromLonLat(lon, lat)
		if err != nil {
			return nil, err
		}

		// Adjust the nonParams array for further operations
		nonParams = nonParams[2:]
	}

	// Fetch Longitute and Latitude from member
	if params[FROMMEMBER] != "" {
		if len(nonParams) < 1 {
			return nil, errors.ErrWrongArgumentCount("GEOSEARCH")
		}
		member := nonParams[0]
		node := geoReg.GetByKey(member)
		if node == nil {
			return nil, errors.ErrMemberNotFoundInSortedSet(member)
		}
		hash := node.Score()
		centerCoordinate, err = NewGeoCoordinateFromHash(uint64(hash))
		if err != nil {
			return nil, err
		}

		// Adjust the nonParams array for further operations
		nonParams = nonParams[1:]
	}

	// Create shape based on BYBOX or BYRADIUS param
	var searchShape GeoShape

	// Create shape from BYBOX
	if params[BYBOX] != "" {
		if len(nonParams) < 2 {
			return nil, errors.ErrWrongArgumentCount("GEOSEARCH")
		}

		var width, height float64
		var errWidth, errHeight error
		width, errWidth = strconv.ParseFloat(nonParams[0], 10)
		height, errHeight = strconv.ParseFloat(nonParams[1], 10)

		if errWidth != nil || errHeight != nil {
			return nil, errors.ErrInvalidNumberFormat
		}
		if height <= 0 || width <= 0 {
			return nil, errors.ErrGeneral("HEIGHT, WIDTH should be > 0")
		}

		searchShape, _ = GetNewGeoShapeRectangle(width, height, centerCoordinate, unit)

		// Adjust the nonParams array for further operations
		nonParams = nonParams[2:]
	}

	// Create shape from BYRADIUS
	if params[BYRADIUS] != "" {
		if len(nonParams) < 1 {
			return nil, errors.ErrWrongArgumentCount("GEOSEARCH")
		}

		var radius float64
		var errRad error
		radius, errRad = strconv.ParseFloat(nonParams[0], 10)

		if errRad != nil {
			return nil, errors.ErrInvalidNumberFormat
		}
		if radius <= 0 {
			return nil, errors.ErrGeneral("RADIUS should be > 0")
		}

		searchShape, _ = GetNewGeoShapeCircle(radius, centerCoordinate, unit)

		// Adjust the nonParams array for further operations
		nonParams = nonParams[1:]
	}

	// Get COUNT based on Params
	var count int = -1
	var errCount error

	// Check for COUNT to limit the output
	if params[COUNT] != "" {
		if len(nonParams) < 1 {
			return nil, errors.ErrWrongArgumentCount("GEOSEARCH")
		}
		count, errCount = strconv.Atoi(nonParams[0])
		if errCount != nil {
			return nil, errors.ErrInvalidNumberFormat
		}
		if count <= 0 {
			return nil, errors.ErrGeneral("COUNT must be > 0")
		}

		// Adjust the nonParams array for further operations
		nonParams = nonParams[1:]
	}

	// If all the params are not used till now
	// Means there're some unknown param
	if len(nonParams) != 0 {
		return nil, errors.ErrUnknownOption(nonParams[0])
	}

	// Check for ANY option
	var anyOption bool = false
	if params[ANY] != "" {
		anyOption = true
	}

	// Check for Sorting Key ASC or DESC (-1 = DESC, 0 = NoSort, 1 = ASC)
	var sortType float64 = 0
	if params[ASC] != "" {
		sortType = 1
	}
	if params[DESC] != "" {
		sortType = -1
	}

	// COUNT without ordering does not make much sense (we need to sort in order to return the closest N entries)
	// Note that this is not needed for ANY option
	if count != -1 && sortType == 0 && !anyOption {
		sortType = 1
	}

	var withCoord, withDist, withHash bool = false, false, false

	if params[WITHCOORD] != "" {
		withCoord = true
	}
	if params[WITHDIST] != "" {
		withDist = true
	}
	if params[WITHHASH] != "" {
		withHash = true
	}

	// Find Neighbors from the shape
	boudingBox := GetBoundingBoxForShape(searchShape)
	neighbors := GetGeohashNeighborsWithinShape(searchShape, boudingBox)
	neighborsArr := neighbors.ToArray()

	// HashMap of all the nodes (we are making map for deduplication)
	geoElementMap := map[string]*wire.GEOElement{}
	totalElements := 0

	// Find all the elements in the neighbor and the center block
	for _, neighbor := range neighborsArr {

		// Discarded neighbors
		if neighbor == 0 {
			continue
		}

		// If ANY option is used and totalElements == count
		// Break the loop and Return the current result
		if anyOption && count == totalElements {
			break
		}

		maxHash, minHash := GetMaxAndMinHashForBoxHash(neighbor, neighbors.Precision)

		zElements := geoReg.ZRANGE(int(minHash), int(maxHash), true, false)

		for _, ele := range zElements {

			if anyOption && totalElements == count {
				break
			}

			eleCoord, _ := NewGeoCoordinateFromHash(uint64(ele.Score))
			dist := searchShape.GetDistanceIfWithinShape(eleCoord)

			if dist != 0 {
				geoElement := wire.GEOElement{
					Member: ele.Member,
					Coordinates: &wire.GEOCoordinates{
						Longitude: centerCoordinate.Longitude,
						Latitude:  centerCoordinate.Latitude,
					},
					Distance: dist,
					Hash:     uint64(ele.Score),
				}
				geoElementMap[ele.Member] = &geoElement
				totalElements++
			}

		}
	}

	// Convert map to array
	geoElements := []*wire.GEOElement{}

	for _, ele := range geoElementMap {
		geoElements = append(geoElements, ele)
	}

	// Return unsorted result if ANY is used or sortType = 0
	if anyOption || sortType == 0 {
		filterDimensionsBasedOnFlags(geoElements, withCoord, withDist, withHash)
		return geoElements, nil
	}

	// Let count be the total elements we need
	if count == -1 {
		count = totalElements
	}

	// Comparator function for MaxHeap
	// If ASC is set -> we use MaxHeap -> To Pop out the largest element if LEN > COUNT
	// If DESC is set -> we use MinHeap -> To Pop out the smallest element if LEN > COUNT
	// So Reverse the final array
	cmp := func(a, b interface{}) int {
		distance1 := a.(*wire.GEOElement).Distance
		distance2 := b.(*wire.GEOElement).Distance
		if distance1*sortType < distance2*sortType {
			return 1
		} else if distance1*sortType > distance2*sortType {
			return -1
		}
		return 0
	}

	// Create a priority Queue to store the 'COUNT' results
	pq := priorityqueue.NewWith(cmp)

	for _, ele := range geoElements {
		pq.Enqueue(ele)
		if pq.Size() > count {
			pq.Dequeue()
		}
	}

	// Final result Arr
	resultGeoElements := []*wire.GEOElement{}

	// Transfer elements from priority Queue to Arr
	for pq.Size() > 0 {
		queueEle, _ := pq.Dequeue()
		geoEle := queueEle.(*wire.GEOElement)
		resultGeoElements = append(resultGeoElements, geoEle)
	}

	// Reverse the output array Because
	// If ASC is set -> we use MaxHeap -> Which will give use DESC array
	// If DESC is set -> we use MinHeap -> Which will give use ASC array
	// So Reverse the final array
	for i, j := 0, len(resultGeoElements)-1; i < j; i, j = i+1, j-1 {
		resultGeoElements[i], resultGeoElements[j] = resultGeoElements[j], resultGeoElements[i]
	}
	filterDimensionsBasedOnFlags(resultGeoElements, withCoord, withDist, withHash)

	return resultGeoElements, nil
}

// This returns 11 characters geohash representation of the position of the specified elements.
func (geoReg *GeoRegistry) Get11BytesHash(members []string) []string {

	result := make([]string, len(members))
	geoAlphabet := []rune("0123456789bcdefghjkmnpqrstuvwxyz")

	for idx, member := range members {

		// Get GEOHASH of the member
		hashNode := geoReg.GetByKey(member)
		if hashNode == nil {
			continue
		}
		hash := uint64(hashNode.Score())

		// Convert the hash to 11 character string (base32)
		hashRune := make([]rune, 11)
		for i := 0; i < 11; i++ {
			var idx uint64
			if i == 10 {
				idx = 0 // pad last char due to only 52 bits
			} else {
				shift := 52 - ((i + 1) * 5)
				idx = (hash >> shift) & 0x1F
			}
			hashRune[i] = geoAlphabet[idx]
		}
		result[idx] = string(hashRune)

	}
	return result
}

// This returns coordinates (longitute, latitude) of all the members
func (geoReg *GeoRegistry) GetCoordinates(members []string) []*GeoCoordinate {
	result := make([]*GeoCoordinate, len(members))
	for i, member := range members {
		// Get GEOHASH of the member
		hashNode := geoReg.GetByKey(member)
		if hashNode == nil {
			continue
		}
		hash := uint64(hashNode.Score())
		coordinate, _ := NewGeoCoordinateFromHash(hash)
		result[i] = coordinate
	}
	return result
}

// /////////////////////////////////////////////////////
// //////////// Utility Functions //////////////////////
// /////////////////////////////////////////////////////

// Filter dimensions based on flags for GEO Search
func filterDimensionsBasedOnFlags(geoElements []*wire.GEOElement, withCoord, withDist, withHash bool) {
	for _, ele := range geoElements {
		if !withCoord {
			ele.Coordinates = nil
		}

		if !withDist {
			ele.Distance = 0
		}

		if !withHash {
			ele.Hash = 0
		}
	}
}

// Encode given Lon and Lat to GEOHASH
func EncodeHash(longitude, latitude float64) uint64 {
	return geohash.EncodeIntWithPrecision(latitude, longitude, BIT_PRECISION)
}

// DecodeHash returns the latitude and longitude from a geo hash
func DecodeHash(hash uint64) (lon, lat float64) {
	lat, lon = geohash.DecodeIntWithPrecision(hash, BIT_PRECISION)
	return lon, lat
}

func ValidateLonLat(lon, lat float64) error {
	if lat > LAT_MAX || lat < LAT_MIN || lon > LONG_MAX || lon < LONG_MIN {
		return errors.ErrInvalidLonLatPair(lon, lat)
	}
	return nil
}

func GetUnitTypeFromParsedParams(params map[Param]string) Param {
	if params[M] != "" {
		return M
	} else if params[KM] != "" {
		return KM
	} else if params[MI] != "" {
		return MI
	} else if params[FT] != "" {
		return FT
	} else {
		return ""
	}
}

func DegToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func RadToDeg(rad float64) float64 {
	return 180.0 * rad / math.Pi
}

// ConvertDistance converts a distance from meters to the desired unit
func ConvertDistance(distance float64, unit Param) (float64, error) {
	var result float64

	switch unit {
	case M:
		result = distance
	case KM:
		result = distance / 1000
	case MI:
		result = distance / 1609.34
	case FT:
		result = distance / 0.3048
	default:
		return 0, errors.ErrInvalidUnit(string(unit))
	}

	// Round to 5 decimal places
	return math.Round(result*10000) / 10000, nil
}

// ConvertToMeter converts a distance to meters from the given unit
func ConvertToMeter(distance float64, unit Param) (float64, error) {
	var result float64

	switch unit {
	case M:
		result = distance
	case KM:
		result = distance * 1000
	case MI:
		result = distance * 1609.34
	case FT:
		result = distance * 0.3048
	default:
		return 0, errors.ErrInvalidUnit(string(unit))
	}
	// Round to 5 decimal places
	return math.Round(result*10000) / 10000, nil
}

// Computes the min (inclusive) and max (exclusive) scores for a given hash box.
// Aligns the geohash bits to BIT_PRECISION score by left-shifting
func GetMaxAndMinHashForBoxHash(hash uint64, precision uint) (max, min uint64) {
	shift := BIT_PRECISION - (precision)
	base := hash << shift
	rangeSize := uint64(1) << shift
	min = base
	max = base + rangeSize - 1

	return max, min
}
