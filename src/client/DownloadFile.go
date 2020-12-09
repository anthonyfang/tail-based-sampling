package client

import (
    "net/http"
    "strconv"
    "sync"
    "os"
    "io"
    "log"
)

var wg sync.WaitGroup

// DownloadFile is download file with Range header
func DownloadFile(url string){
    res, _ := http.Head(url);
    maps := res.Header
    length, _ := strconv.Atoi(maps["Content-Length"][0])
    limit := 100
    lenSub := length / limit
    diff := length % limit
    for i := 0; i < limit ; i++ {
        wg.Add(1)

        min := lenSub * i // Min range
        max := lenSub * (i + 1) // Max range

        if (i == limit - 1) {
            max += diff // Add the remaining bytes in the last request
        }

        go func(min int, max int, i int) {
            client := &http.Client {}
            req, _ := http.NewRequest("GET", url, nil)  
            rangeHeader := "bytes=" + strconv.Itoa(min) +"-" + strconv.Itoa(max-1) // Add the data for the Range header of the form "bytes=0-100"
            req.Header.Add("Range", rangeHeader)
            resp,_ := client.Do(req)
            defer func ()  { resp.Body.Close() }()

            // Create the file
            f, _ := os.Create("/tmp/"+ strconv.Itoa(i))
            defer func ()  {
                f.Close()
                wg.Done()
            }()

            // Write the body to file
            _, _ = io.Copy(f, resp.Body)
        }(min, max, i)
    }
    wg.Wait()

    // Merge files
    for i:=0; i < limit ; i++ {
        fileMerge(strconv.Itoa(i))
    }
    downloadChan <- true
}

func fileMerge(fileName string) {
    srcFilePath := "/tmp/datafile"
    dstFilePath := "/tmp/" + fileName

    fileACopy, err := os.Open(dstFilePath)
    if err != nil {
        log.Fatal(err)
    }
    defer fileACopy.Close()

    append, err := os.OpenFile(srcFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer append.Close()

    io.Copy(append, fileACopy)
    _ = os.Remove(dstFilePath)
}
