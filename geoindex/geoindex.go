package geoindex

import (
	"fmt"
	"os"
	"sync"
	"sort"
	"bufio"
	"strings"
	"strconv"
	"math"
	//"time"
)

//////////////////////
// Global Constants //
//////////////////////
const Earth_Radius = 6378.1 //km


//////////////////////////////
// Custom type for sorting  //
//////////////////////////////
type intArray []int64
func (s intArray) Len() int { return len(s) }
func (s intArray) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s intArray) Less(i, j int) bool { return s[i] < s[j] }


///////////////////////
// Geo lookup object //
///////////////////////
type IndexedGeoObj struct {
	sync.RWMutex
	M map[int64]string
	I []int64
}

func (Io *IndexedGeoObj) Locate(lat float64, lon float64, sample int) string{ //, sample int
	hash, _ := Encode_uint64(lat, lon)

	// get desired candidate points
	index := float64(binarysearch(Io.I, hash))

	candidates := make([]int64 , sample)
	j := 0
	for i:=math.Max(0, index-float64(sample)/2); i<math.Min(float64(len(Io.I)), index+float64(sample)/2); i++ {
		candidates[j] = Io.I[int(i)]
		j++
	}
	
	var result int64
	min_dist := 4.0 //  range of arccos is [0,pi], set this to 4 to represent +inf
	for i:=0; i<sample; i++ {
		lat1, lon1 := Decode_uint64(uint64(candidates[i]))

		dist := Distance_on_unit_sphere(lat, lon, lat1, lon1)
		
		if dist < min_dist {
			result = candidates[i]
			min_dist = dist
		}
	}

	return Io.M[result]
}

func LoadGeoFromFile(fileloc string, geoobj *IndexedGeoObj) error{
	// read from the file
	fp, err := os.Open(fileloc)
	if err != nil {
		return fmt.Errorf("Unable to open Geo file\n  %s", err)
	}
	defer fp.Close()

	geoobj.Lock()

	// index each point and add them to the Object
	geoobj.M = make(map[int64]string)

	s := bufio.NewScanner(fp)
	var line string
	for s.Scan() {
		line = s.Text()
		if line != ""{
			content := strings.Split(line, ",")
			pcode  := content[0]
			lat, _ := strconv.ParseFloat(content[1], 64)
			lon, _ := strconv.ParseFloat(content[2], 64)

			geoi, _ := Encode_uint64(lat, lon)

			geoobj.M[geoi] = pcode
		}
	}

	// create the index list
	indexmax := len(geoobj.M)
	geoobj.I = make([]int64, indexmax)

	i := 0
	for ind,_ := range geoobj.M {
		geoobj.I[i] = ind
		i++
	}
	// sort it
	sort.Sort(intArray(geoobj.I))

	geoobj.Unlock()
	return nil
}


////////////////////
// Encoding funcs //
////////////////////
func Encode_uint64(latitude float64, longitude float64) (int64, error){
	if latitude >= 90.0 || latitude < -90.0 {
		return -1, fmt.Errorf("Latitude must be in the range of (-90.0, 90.0)")
	}
	for longitude < -180.0 {
		longitude += 360.0
	}
	for longitude >= 180.0 {
		longitude -= 360.0
	}

	lat := int64(((latitude + 90.0)/180.0)*(1<<32))
	lon := int64(((longitude+180.0)/360.0)*(1<<32))
	return _uint64_interleave(lat, lon), nil
}

func _uint64_interleave(lat32 int64, lon32 int64) int64{
	intr := int64(0)
	boost := []int64{0,1,4,5,16,17,20,21,64,65,68,69,80,81,84,85}

	var shift uint64
	for i:=0; i<8; i++ {
		shift = uint64(28-i*4)
		intr = (intr<<8) + (boost[(lon32>>shift)%16]<<1) + (boost[(lat32>>shift)%16])
	}
	return intr
}

func Decode_uint64(ui64 uint64) (float64, float64) {
	lat, lon := _uint64_deinterleave(ui64)

	return float64(180.0*lat)/float64(1<<32) - 90.0, float64(360.0*lon)/float64(1<<32) - 180.0
}

func _uint64_deinterleave(ui64 uint64) (int64, int64){
	lat := int64(0)
	lon := int64(0)

	boost := [][2]int64{ [2]int64{0,0},[2]int64{0,1},[2]int64{1,0},[2]int64{1,1}, [2]int64{0,2}, [2]int64{0,3},[2]int64{1,2},[2]int64{1,3},[2]int64{2,0},[2]int64{2,1},[2]int64{3,0},[2]int64{3,1},[2]int64{2,2},[2]int64{2,3},[2]int64{3,2},[2]int64{3,3}}

	var shift uint64
	var p [2]int64
	for i:=0; i<16; i++ {
		shift = uint64(60-i*4)

		p = boost[(ui64>>shift)%16]

		lat = (lat<<2) + p[1]
		lon = (lon<<2) + p[0]
	}
	return lat, lon
}


//////////////////
// Helper funcs //
//////////////////
func midpoint(lo int64, hi int64) int64{
	return lo + ((hi - lo) / 2)
}

func Binarysearch(l []int64, target int64) int64{
	var imin, imid, imax int64

	imin = 0
	imax = int64(len(l)-1)

	for imax > imin {
		imid = midpoint(int64(imin), int64(imax)) 

		if imid >= imax {
			//fmt.Println("12345")
			break
		}

		if l[imid] == target {
			return imid
		} else if l[imid] < target {
			imin = imid + 1
		} else{
			imax = imid
		}
	}

	return imin
}

func Distance_on_unit_sphere(lat1 float64, long1 float64, lat2 float64, long2 float64) float64{

	degrees_to_radians := math.Pi / 180.0

	// phi = 90 - latitude
	phi1 := (90.0 - lat1)*degrees_to_radians
	phi2 := (90.0 - lat2)*degrees_to_radians
	
	// theta = longitude
	theta1 := long1*degrees_to_radians
	theta2 := long2*degrees_to_radians


	// Compute spherical distance from spherical coordinates.

	// For two locations in spherical coordinates 
	// (1, theta, phi) and (1, theta, phi)
	// cosine( arc length ) = 
	//  sin phi sin phi' cos(theta-theta') + cos phi cos phi'
	// distance = rho * arc length

	cos := (math.Sin(phi1)*math.Sin(phi2)*math.Cos(theta1 - theta2) + math.Cos(phi1)*math.Cos(phi2))
	arc := math.Acos( cos )

	return arc
}
