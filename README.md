rgogeo
-

a reverse geocoding lib in go

#### Install
```
go get github.com/frankyin1019/rgeogo
```

#### Format of data files
This lib includes 2 data files under `data` folder that covers US and Canada zip/postal codes. 
The accuracy is NOT guranteed. 
You should always use any data sources of your choice. 

To feed in the data:
- name the file as COUNTRY.csv
- fields of the csv: `postal code, lat, lon, city, region`
- put all the data files in a folder
- pass the path to the folder to `rgeogo.Setup(path)`

#### Sample Use Case
```
import (
	"fmt"
	"github.com/frankyin1019/rgeogo"
) 

func main() {
	rgeogo.Setup("/PATH/TO/DATA/FOLDER")

	fmt.Println(rgeogo.RGeocode(LAT, LON, 8))
}
```

