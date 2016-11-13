package main

import (
        "fmt"
        "image"
        _ "image/gif"
        _ "image/jpeg"
        _ "image/png"
        "log"
        "net/http"
)

// taken from https://gist.github.com/tristanwietsma/c552e838f21f6fbb5800
func calcHistogram(url string) *[16][4]int {
        fmt.Println("image: calcHistogram: url=", url)

        req, err := http.NewRequest("GET", url, nil)
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
                panic(err)
        }
        defer resp.Body.Close()

        fmt.Println("image: calcHistogram: response Status:", resp.Status)

        m, _, err := image.Decode(resp.Body)
        if err != nil {
                log.Fatal(err)
        }
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

func main() {
        histogram := calcHistogram("http://im.ft-static.com/content/images/2b173ba4-ca7c-4d32-8478-7a2094437eeb.img")
        // Print the results.
        fmt.Printf("%-14s %6s %6s %6s %6s\n", "bin", "red", "green", "blue", "alpha")
        for i, x := range histogram {
                fmt.Printf("0x%04x-0x%04x: %6d %6d %6d %6d\n", i<<12, (i+1)<<12-1, x[0], x[1], x[2], x[3])
        }
}