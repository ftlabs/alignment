package image

import (
        "fmt"
        "image"
        _ "image/gif"
        _ "image/jpeg"
        _ "image/png"
        "log"
        "net/http"
        "sort"
        "github.com/generaltso/vibrant"
)

var checkErr = func(err error) { 
        if err != nil { 
                panic(err) 
                log.Fatal(err)
                } 
        }

func getDecodedImageByUrl(url string) *image.Image {
        fmt.Println("image: getDecodedImageByUrl: url=", url)

        req, err := http.NewRequest("GET", url, nil)
        client := &http.Client{}
        resp, err := client.Do(req)
        checkErr( err )
        defer resp.Body.Close()

        fmt.Println("image: getDecodedImageByUrl: response Status:", resp.Status)

        m, _, err := image.Decode(resp.Body)
        checkErr( err )
        return &m
}

// taken from https://gist.github.com/tristanwietsma/c552e838f21f6fbb5800
func calcHistogram(url string) *[16][4]int {
        m := *getDecodedImageByUrl( url )
        bounds := m.Bounds()

        var histogram [16][4]int
        for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
                for x := bounds.Min.X; x < bounds.Max.X; x++ {
                        r, g, b, a := m.At(x, y).RGBA()
                        // A color's RGBA method returns values in the range [0, 65535].
                        // Shifting by 12 reduces this to the range [0, 15].
                        histogram[r>>12][0]++
                        histogram[g>>12][1]++
                        histogram[b>>12][2]++
                        histogram[a>>12][3]++
                }
        }

        return &histogram
}

type ColourStat struct {
        RgbaCsv     string
        Count       int
        Percentage  float64
}

type ByCount []ColourStat

func (s ByCount) Len() int {
        return len(s)
}
func (s ByCount) Swap(i, j int) {
        s[i], s[j] = s[j], s[i]
}
func (s ByCount) Less(i, j int) bool {
        return s[i].Count > s[j].Count
}

func calcColourFrequencies(url string) *[]ColourStat {
        m := *getDecodedImageByUrl( url )
        bounds := m.Bounds()

        var colourCounts = make(map[string]int)
        var absCount = 0

        for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
                for x := bounds.Min.X; x < bounds.Max.X; x++ {
                        r, g, b, a := m.At(x, y).RGBA()
                        rgbaString := fmt.Sprintf( "%03d,%03d,%03d,%03d",
                                (r>>12)<<4,
                                (g>>12)<<4,
                                (b>>12)<<4,
                                (a>>12)<<4,
                                )
                        colourCounts[rgbaString]++
                        absCount ++
                }
        }

        colourStats := []ColourStat {}        
        for k,v := range colourCounts {
                colourStat := ColourStat{
                      RgbaCsv: k,
                      Count:   v,
                      Percentage: (float64(v) * 100.) / float64(absCount),
                }
                colourStats = append( colourStats, colourStat )
        }

        sort.Sort(ByCount(colourStats))


        return &colourStats
}

type ProminentColour struct {
        Name       string
        Population int
        RGBHex     string
}

type ByPopulation []ProminentColour

func (s ByPopulation) Len() int {
        return len(s)
}
func (s ByPopulation) Swap(i, j int) {
        s[i], s[j] = s[j], s[i]
}
func (s ByPopulation) Less(i, j int) bool {
        return s[i].Population > s[j].Population
}

var imgProminentColoursCache = map[string]*[]ProminentColour{}

// via https://github.com/generaltso/vibrant
func GetProminentColours(url string) *[]ProminentColour {
    var prominentColours *[]ProminentColour
    
    if _, ok := imgProminentColoursCache[url]; ok {
        fmt.Println("image.GetProminentColours: cache hit: url=", url)
        prominentColours = imgProminentColoursCache[url]
    } else {
        fmt.Println("image.GetProminentColours: cache miss: url=", url)

        prominentColours = &([]ProminentColour {})
        img := *getDecodedImageByUrl( url )

        palette, err := vibrant.NewPaletteFromImage(img)
        checkErr(err)

        swatches := palette.ExtractAwesome()

        for name, swatch := range swatches {
          // fmt.Printf("/* %s (population: %d) */\n%s\n\n", name, swatch.Population, swatch)
          prominentColour := ProminentColour{
                Name:       name,
                Population: swatch.Population,
                RGBHex:     swatch.Color.RGBHex(),
          }

          *prominentColours = append( *prominentColours, prominentColour )
        }

        sort.Sort(ByPopulation(*prominentColours))

        imgProminentColoursCache[url] = prominentColours
    }

    return prominentColours
}

func main() {
        url := "http://im.ft-static.com/content/images/2a7f93c1-0276-4e3e-992d-d14a60bf60f4.img"

        // histogram := calcHistogram(url)
        // // Print the results.
        // fmt.Printf("%-14s %6s %6s %6s %6s\n", "bin", "red", "green", "blue", "alpha")
        // for i, x := range histogram {
        //         fmt.Printf("0x%04x-0x%04x: %6d %6d %6d %6d\n", i<<12, (i+1)<<12-1, x[0], x[1], x[2], x[3])
        // }

        // colourStats := calcColourFrequencies( url )
        // for i,c := range *colourStats{
        //         fmt.Printf("%3d) %s %5.2f %6d\n", i, c.RgbaCsv, c.Percentage, c.Count)
        // }

        prominentColours1 := GetProminentColours( url )
        fmt.Println( "image: main: prominentColours1=", prominentColours1)      
        prominentColours2 := GetProminentColours( url )
        fmt.Println( "image: main: prominentColours2=", prominentColours2)      
}