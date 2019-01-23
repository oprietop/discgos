// https://github.com/brotherlogic/godiscogs/blob/master/DiscogsRetriever.go
package main

import (
    "flag"
    "fmt"
    "time"
    "os"
    "io"
    "net/http/httputil"
    "regexp"
    "bufio"
    "bytes"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"
    "sort"
    "path/filepath"
    "syscall"                     // change the UID
    "github.com/mewkiz/flac"      // handle Flac Streams
    "github.com/mewkiz/flac/meta" // handle Vorbis Tags
)

// key and secret from https://github.com/beetbox/beets/blob/master/beetsplug/discogs.py
const (
    key    = "rAzVUQYRaoFjeBjyWuWZ"
    secret = "plxtUTqoCzwxZpqdPysCwGuBSmZNdZVy"
    ua     = "net/http Golang"
)

// Map with default api args for each call
var args = map[string]string{
    "per_page": "100",
}

// http://json2struct.mervine.net/
type Pagination struct {
    Items   int      `json:"items"`
    Page    int      `json:"page"`
    Pages   int      `json:"pages"`
    PerPage int      `json:"per_page"`
    Urls    struct {
        Last string `json:"last"`
        Next string `json:"next"`
    } `json:"urls"`
}

type Search struct {
    Pagination Pagination  `json:"pagination"`
    Results []struct {
        Barcode   []string `json:"barcode"`
        Catno     string   `json:"catno"`
        Community struct {
            Have int `json:"have"`
            Want int `json:"want"`
        } `json:"community"`
        Country     string   `json:"country"`
        Format      []string `json:"format"`
        Genre       []string `json:"genre"`
        ID          int      `json:"id"`
        Label       []string `json:"label"`
        ResourceURL string   `json:"resource_url"`
        Style       []string `json:"style"`
        Thumb       string   `json:"thumb"`
        Title       string   `json:"title"`
        Type        string   `json:"type"`
        URI         string   `json:"uri"`
        Year        string   `json:"year"`
    } `json:"results"`
}

type MasterRelease struct {
    Artists []struct {
        Anv         string `json:"anv"`
        ID          int    `json:"id"`
        Join        string `json:"join"`
        Name        string `json:"name"`
        ResourceURL string `json:"resource_url"`
        Role        string `json:"role"`
        Tracks      string `json:"tracks"`
    } `json:"artists"`
    DataQuality string   `json:"data_quality"`
    Genres      []string `json:"genres"`
    ID          int      `json:"id"`
    Images      []struct {
        Height      int    `json:"height"`
        ResourceURL string `json:"resource_url"`
        Type        string `json:"type"`
        URI         string `json:"uri"`
        URI150      string `json:"uri150"`
        Width       int    `json:"width"`
    } `json:"images"`
    LowestPrice    interface{} `json:"lowest_price"`
    MainRelease    int         `json:"main_release"`
    MainReleaseURL string      `json:"main_release_url"`
    NumForSale     int         `json:"num_for_sale"`
    ResourceURL    string      `json:"resource_url"`
    Styles         []string    `json:"styles"`
    Title          string      `json:"title"`
    Tracklist      []struct {
        Duration string `json:"duration"`
        Position string `json:"position"`
        Title    string `json:"title"`
        Type     string `json:"type_"`
    } `json:"tracklist"`
    URI         string `json:"uri"`
    VersionsURL string `json:"versions_url"`
    Year        int    `json:"year"`
}

type Release struct {
    Artists []struct {
        Anv         string `json:"anv"`
        ID          int    `json:"id"`
        Join        string `json:"join"`
        Name        string `json:"name"`
        ResourceURL string `json:"resource_url"`
        Role        string `json:"role"`
        Tracks      string `json:"tracks"`
    } `json:"artists"`
    Community struct {
        Contributors []struct {
            ResourceURL string `json:"resource_url"`
            Username    string `json:"username"`
        } `json:"contributors"`
        DataQuality string `json:"data_quality"`
        Have        int    `json:"have"`
        Rating      struct {
            Average float64 `json:"average"`
            Count   int `json:"count"`
        } `json:"rating"`
        Status    string `json:"status"`
        Submitter struct {
            ResourceURL string `json:"resource_url"`
            Username    string `json:"username"`
        } `json:"submitter"`
        Want int `json:"want"`
    } `json:"community"`
    Companies []struct {
        Catno          string `json:"catno"`
        EntityType     string `json:"entity_type"`
        EntityTypeName string `json:"entity_type_name"`
        ID             int    `json:"id"`
        Name           string `json:"name"`
        ResourceURL    string `json:"resource_url"`
    } `json:"companies"`
    Country         string `json:"country"`
    DataQuality     string `json:"data_quality"`
    DateAdded       string `json:"date_added"`
    DateChanged     string `json:"date_changed"`
    EstimatedWeight int    `json:"estimated_weight"`
    Extraartists    []struct {
        Anv         string `json:"anv"`
        ID          int    `json:"id"`
        Join        string `json:"join"`
        Name        string `json:"name"`
        ResourceURL string `json:"resource_url"`
        Role        string `json:"role"`
        Tracks      string `json:"tracks"`
    } `json:"extraartists"`
    FormatQuantity int `json:"format_quantity"`
    Formats        []struct {
        Descriptions []string `json:"descriptions"`
        Name         string   `json:"name"`
        Qty          string   `json:"qty"`
    } `json:"formats"`
    Genres      []string `json:"genres"`
    ID          int      `json:"id"`
    Identifiers []struct {
        Type  string `json:"type"`
        Value string `json:"value"`
    } `json:"identifiers"`
    Images []struct {
        Height      int    `json:"height"`
        ResourceURL string `json:"resource_url"`
        Type        string `json:"type"`
        URI         string `json:"uri"`
        URI150      string `json:"uri150"`
        Width       int    `json:"width"`
    } `json:"images"`
    Labels []struct {
        Catno          string `json:"catno"`
        EntityType     string `json:"entity_type"`
        EntityTypeName string `json:"entity_type_name"`
        ID             int    `json:"id"`
        Name           string `json:"name"`
        ResourceURL    string `json:"resource_url"`
    } `json:"labels"`
    LowestPrice       float64       `json:"lowest_price"`
    MasterID          int           `json:"master_id"`
    MasterURL         string        `json:"master_url"`
    Notes             string        `json:"notes"`
    NumForSale        int           `json:"num_for_sale"`
    Released          string        `json:"released"`
    ReleasedFormatted string        `json:"released_formatted"`
    ResourceURL       string        `json:"resource_url"`
    Series            []interface{} `json:"series"`
    Status            string        `json:"status"`
    Styles            []string      `json:"styles"`
    Thumb             string        `json:"thumb"`
    Title             string        `json:"title"`
    Tracklist         []struct {
        Duration string `json:"duration"`
        Position string `json:"position"`
        Title    string `json:"title"`
        Type     string `json:"type_"`
    } `json:"tracklist"`
    URI    string `json:"uri"`
    Videos []struct {
        Description string `json:"description"`
        Duration    int    `json:"duration"`
        Embed       bool   `json:"embed"`
        Title       string `json:"title"`
        URI         string `json:"uri"`
    } `json:"videos"`
    Year int `json:"year"`
}

type Version struct {
    Catno        string   `json:"catno"`
    Country      string   `json:"country"`
    Format       string   `json:"format"`
    ID           int      `json:"id"`
    MasterID     int      // Custom!
    Label        string   `json:"label"`
    MajorFormats []string `json:"major_formats"`
    Released     string   `json:"released"`
    ResourceURL  string   `json:"resource_url"`
    Status       string   `json:"status"`
    Thumb        string   `json:"thumb"`
    Title        string   `json:"title"`
}

// Slice of Versions
type Versions []*Version

// Implement a custom sort.Interface to sort by values
func (v Versions) Len() int           { return len(v) }
func (v Versions) Swap(a, b int)      { v[a], v[b] = v[b], v[a] }

// Our sort preference will be Released > Country > Id
func (v Versions) Less(a, b int) bool {
    // Check for proper release dates (4 digits for a year)
    if len(v[a].Released) < 4 || len(v[b].Released) < 4 {
        return true
    }

    // Compare only the first four digits
    if v[a].Released[0:4] < v[b].Released[0:4] {
        return true
    } else if v[a].Released[0:4] > v[b].Released[0:4] {
        return false
    }

    // If we're here it means the Release year was the same
    if v[a].Country < v[b].Country {
        return true
    } else if v[a].Country > v[b].Country {
        return false
    }

    // If we're here it means the Country was also the same
    return v[a].ID < v[b].ID
}

type MasterVersions struct {
    Pagination Pagination  `json:"pagination"`
    Versions   Versions    `json:"versions"`
}

type Results struct {
    MasterReleases map[int]*MasterRelease
    MasterVersions map[int]*Version
}

type Flags struct {
    Extra    *string
    Regexp   *string
    Debug    *bool
    Id       *int
    Uid      *int
}

type Query struct {
    Master       *MasterRelease
    Release      *Release
    Flags        Flags
    Query        string
    OldDir       string
    CleanOldDir  string
    FlacFiles    []string
    OtherFiles   []string
    NewAlbumDir  string
    MediaName    string
    Tags         map[string]string
}

type Discogs struct {
    auth      string
    ua        string
    Flags     Flags
}

// Error format helper
func check(err error) {
    if err != nil && err != io.EOF {
        fmt.Printf("# Error: %s\n", c(1, 31, err))
    }
}

// Naive ANSII Color Helper
func c(a, b int, i interface{}) string {
    return fmt.Sprintf("\033[%d;%dm%v\033[0m", a, b, i)
}

// Debug Helper
func d(is ...interface{}) {
    for _, i := range is {
        b, err := json.MarshalIndent(i, "", "  ")
        // check ir error or b is holding an empty struct "{}" (2 bytes)
        if err != nil || len(b) == 2 {
            fmt.Printf("%#v\n", i)
        } else {
            fmt.Printf("%s\n", b)
        }
    }
}

// Remove unwanted parts of a string
func cleanString(s string) string {
    comments := regexp.MustCompile(`[\[\{\(]+.*?[\)\}\]]+`)
    s = comments.ReplaceAllString(s, "")
    s = strings.Replace(s, "/", "", -1)
    return strings.Trim(s, " ")
}

// Try a filename friendlier to some filesystems
func cleanFilename(s string) string {
    unwantedChars := regexp.MustCompile(`[\x5C\r\n\t\013/|*"?<>:]`)
    return unwantedChars.ReplaceAllString(s, "-")
}

// Fill the tag map
func buildTags(q *Query) {
    // 1 to 1 tag assignments
    //d(q.Release)
    q.Tags["RELEASEDATE"]    = q.Release.Released
    q.Tags["MEDIA_NUM"]      = strconv.Itoa(q.Release.FormatQuantity)
    q.Tags["MEDIA"]          = q.Release.Formats[0].Name
    q.Tags["LABEL"]          = q.Release.Labels[0].Name
    q.Tags["CATALOGNUMBER"]  = q.Release.Labels[0].Catno
    q.Tags["RELEASECOUNTRY"] = q.Release.Country
    q.Tags["GENRE"]          = q.Release.Genres[0]
    q.Tags["NOTES"]          = q.Release.Notes
    q.Tags["ALBUMNAME"]      = q.Release.Title
    q.Tags["DATE"]           = strconv.Itoa(q.Master.Year)
    q.Tags["ORIGINALDATE"]   = strconv.Itoa(q.Master.Year)

    // Use the Artist from the Master, fallback to Release
    if len(q.Master.Artists) > 0 {
        q.Tags["ARTIST"] = cleanString(q.Master.Artists[0].Name)
    } else {
        q.Tags["ARTIST"] = cleanString(q.Release.Artists[0].Name)
    }
    q.Tags["ALBUM ARTIST"] = q.Tags["ARTIST"]

    // Workaround for nasty "0" years :(
    if q.Master.Year == 0 {
       q.Tags["ORIGINALDATE"] = q.Release.Released
       q.Tags["DATE"] = q.Tags["ORIGINALDATE"]
    }

    // Try to use a generic label in this case
    if  strings.HasPrefix(q.Tags["LABEL"], "Not On Label") {
        q.Tags["LABEL"] = "Not On Label"
    }

    // Fallback to genre for the style tag
    if len(q.Release.Styles) > 0 {
        q.Tags["STYLE"] = q.Release.Styles[0]
        q.Tags["STYLES"] = strings.Join(q.Release.Styles, ", ")
    } else {
        q.Tags["STYLE"] = q.Tags["GENRE"]
    }


    // Store the original directory name
    q.Tags["ORIGINAL_DIR"] = filepath.Base(q.OldDir)

    // Discogs related tags (Release)
    q.Tags["DISCOGS_RELEASE_ID"]     = strconv.Itoa(q.Release.ID)
    q.Tags["DISCOGS_RELEASE_URL"]    = "http://www.discogs.com/release/" + strconv.Itoa(q.Release.ID)
    q.Tags["DISCOGS_RELEASE_STATUS"] = q.Release.Status
    if len(q.Release.Images) > 0 {
        q.Tags["DISCOGS_RELEASE_COVER_URL"] = q.Release.Images[0].URI
    }

    // Discogs related tags (Master)
    q.Tags["DISCOGS_MASTER_ID"]  = strconv.Itoa(q.Master.ID)
    q.Tags["DISCOGS_MASTER_URL"] = "http://www.discogs.com/master/view/" + strconv.Itoa(q.Master.ID)
    if len(q.Master.Images) > 0 {
        q.Tags["DISCOGS_MASTER_COVER_URL"]  = q.Master.Images[0].URI
    }
}

// Generate the new album and directory names
func buildName(q *Query) {
    AlbumInfo := []string{}
    // Only want the Release year if it's newer than the original one
    if q.Tags["RELEASEDATE"] > q.Tags["ORIGINALDATE"] {
        AlbumInfo = append(AlbumInfo, q.Tags["RELEASEDATE"])
    }
    AlbumInfo = append(AlbumInfo, cleanString(q.Tags["LABEL"]))
    AlbumInfo = append(AlbumInfo, q.Release.Labels[0].Catno)

    // Add the MediaName string if any
    if q.MediaName != "" {
        AlbumInfo = append(AlbumInfo, q.MediaName)
        q.Tags["MEDIA_NAME"] = q.MediaName
    }

    // Add the country
    AlbumInfo = append(AlbumInfo, q.Tags["RELEASECOUNTRY"])

    // Finally, add Extra info if any
    if *q.Flags.Extra != "" {
        AlbumInfo = append(AlbumInfo, *q.Flags.Extra)
        q.Tags["EXTRA"] = *q.Flags.Extra
    }

    // Store the name and dir for future usage
    q.Tags["ALBUM"] = fmt.Sprintf("%s (%s)", q.Tags["ALBUMNAME"], strings.Join(AlbumInfo," / "))
    q.NewAlbumDir = cleanString(q.Tags["ARTIST"]) + " - " + q.Tags["ORIGINALDATE"] + " - " + q.Tags["ALBUM"]

    // Directory Name must be sanitized
    q.NewAlbumDir = cleanFilename(q.NewAlbumDir)
}

// Get all files with full path from a directory
func FilePathWalkDir(root string) ([]string, error) {
    var files []string
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        // Ignore directories
        if !info.IsDir() {
            files = append(files, path)
        }
        return nil
    })
    return files, err
}

// https://stackoverflow.com/a/21067803
func copyFileContents(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return
    }
    err = out.Sync()

    return
}

func tagAndCopy(q *Query) error {
    files, err := FilePathWalkDir(q.OldDir)
    check(err)

    sort.Strings(files)

    for _, f := range files {
        if strings.HasSuffix(f, ".flac") {
            q.FlacFiles = append(q.FlacFiles, f)
        } else {
            q.OtherFiles = append(q.OtherFiles, f)
        }
    }

    if len(q.FlacFiles) < 1 {
        fmt.Printf("# No flac files\n")
        return fmt.Errorf("# No flac files!")
    }

    // Create a proper tracklist with only the actual tracks
    // Traverse the slice checking his length on each iteration
    for i := len(q.Release.Tracklist) - 1; i >= 0; i-- {
        item := q.Release.Tracklist[i]
        // Check it the item is not of the type "track".
        if item.Type != "track" {
            // If that's the case remove it from the slice
            q.Release.Tracklist = append(q.Release.Tracklist[:i], q.Release.Tracklist[i+1:]...)
        }
    }

    // Warn if we're not being consistent track-wise
    if len(q.Release.Tracklist) != len(q.FlacFiles) {
        fmt.Printf("# Warning, release(%s) and directory(%s) differs in number of tracks!\n", c(1, 33, len(q.Release.Tracklist)), c(1, 33, len(q.FlacFiles)))
        fmt.Printf("Enter to continue...")
        fmt.Scanln()
    }

    // Update the totaltracks tag with our file count
    q.Tags["TOTALTRACKS"] = strconv.Itoa(len(q.FlacFiles))

    // Padding we'll use
    zeroes := len(q.Tags["TOTALTRACKS"])

    // Create the destination directory
    if _, err := os.Stat(q.NewAlbumDir); os.IsNotExist(err) {
        fmt.Printf("# Creating directory '%s'\n", q.NewAlbumDir)
        err := os.Mkdir(q.NewAlbumDir, os.ModePerm)
        if err != nil {
            return fmt.Errorf("# Error creating directory: '%s'", c(1, 31, err))
        }
    } else {
        return fmt.Errorf("# Directory '%s' already exists!", c(1, 31, q.NewAlbumDir))
    }

    // Process the flac files
    for trackNum, f := range q.FlacFiles {
        // Our default destination name for the file
        destFile := filepath.Base(f)
        fmt.Printf("- Tagging '%s', ", c(0, 32, destFile))

        // Decode the source file
        stream, err := flac.ParseFile(f)
        check(err)
        defer stream.Close()

        // Retag the stream
        for _, block := range stream.Blocks {
            if comment, ok := block.Body.(*meta.VorbisComment); ok {
                // newTags will join the current file tags with the new ones
                newTags := map[string]string{}

                // add the current tags
                for _, tag := range comment.Tags {
                    // Tag key to uppercase to avoid duplicates
                    k, v := strings.ToUpper(tag[0]), tag[1]
                    newTags[k] = v
                }

                // Iterate and overlap the new tags over the current ones
                for k, v := range q.Tags {
                    // Just in case
                    k = strings.ToUpper(k)
                    newTags[k] = v
                }

                // Retag Tracknumber/Title if we got the proper info from Discogs
                // First check if we actually got a fitting Tracklist
                if len(q.Release.Tracklist) == len(q.FlacFiles) {
                    // Then check if we got a Title
                    if q.Release.Tracklist[trackNum].Title != "" {
                        newTags["TRACKNUMBER"] = fmt.Sprintf("%0*d", zeroes, trackNum+1)
                        newTags["TITLE"] = q.Release.Tracklist[trackNum].Title
                        destFile = fmt.Sprintf("%s - %s.flac", newTags["TRACKNUMBER"], newTags["TITLE"])
                    }
                }

                // Add stream and file related tags
                newTags["BITSPERSAMPLE"] = strconv.Itoa(int(stream.Info.BitsPerSample))
                newTags["SAMPLERATE"]    = strconv.Itoa(int(stream.Info.SampleRate))

                // Blank the current tags
                comment.Tags = [][2]string{}

                // Append the new tags
                for k, v := range newTags {
                    tag := [2]string{k, v}
                    comment.Tags = append(comment.Tags, tag)
                }
                //comment.Vendor = "foobar"
                //d(comment)
            }
        }

        // Filehandle for the destination file
        destFile = cleanFilename(destFile)
        newFile, err := os.Create(q.NewAlbumDir + "/" + destFile)
        check(err)
        defer newFile.Close()

        // Write the stream
        fmt.Printf("as '%s'\n", c(1, 32, destFile))
        err = flac.Encode(newFile, stream)
        check(err)
    }

    // Process the non flac files
    for _, f := range q.OtherFiles {
        // Remove the base directory
        noBaseFile := strings.Replace(f, q.OldDir + "/", "", -1)
        fmt.Printf("- Copying '%s'\n", c(1, 35, noBaseFile))

        // Split the path from the file
        path, file := filepath.Split(noBaseFile)

        // Append the path to the new album dir
        newPath := filepath.Join(q.NewAlbumDir, path)

        // Create all the path, won't fail if it already exists
        os.MkdirAll(newPath, os.ModePerm)

        // Add the filename to the previous path and do the copy
        newFilePath := filepath.Join(newPath, file)
        copyFileContents(f, newFilePath)
    }

    return nil
}

// Return the file from the binary magic
func getFileType(b []byte) (string) {
    m := http.DetectContentType(b)
    sl := strings.Split(m, "/")

    return sl[len(sl)-1]
}

// Fetch the album cover
func getCover(r *Discogs, q *Query) {
    coverUrl := "https://s.discogs.com/images/default-release-cd.png"
    fmt.Println("# Fetching cover")
    // Try for the release cover
    if len(q.Release.Images) > 0 {
        fmt.Println("- Using Release cover.")
        coverUrl = q.Release.Images[0].URI
    // Fallback to the master one
    } else if len(q.Master.Images) > 0 {
        fmt.Println("- Using Master cover.")
        coverUrl = q.Master.Images[0].URI
    // Bad luck
    } else {
        fmt.Println("- No cover found!")
    }

    // Fetch the cover
    res, err := r.fetch(coverUrl, map[string]string{})
    check(err)

    // Convert to bytes and get the extension from the file's magic
    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Body)
    bytes := buf.Bytes()
    ext := getFileType(bytes)

    // Used to "jpg"
    if ext == "jpeg" {
        ext = "jpg"
    }

    // Create a filehandle to store it
    file, err := os.Create(q.NewAlbumDir + "/" + "folder." + ext)
    check(err)
    defer file.Close()
    // Write the contents, could use res.Body if not using a buffer
    _, err = io.Copy(file, buf)
    check(err)
}

// Initialize a discogs struct with our constants
func NewDiscogs() *Discogs {
    return &Discogs{
        auth: "Discogs key=" + key + ", secret=" + secret,
        ua: ua,
    }
}

func (r *Discogs) fetch(url string, a map[string]string) (*http.Response, error) {
    // Create a GET request
    req, err := http.NewRequest("GET", url, nil)
    check(err)

    // Build our arguments into the query
    q := req.URL.Query()
    for k, v := range a {
        q.Add(k, v)
    }
    req.URL.RawQuery = q.Encode()

    // Set the headers
    req.Header.Set("User-Agent", r.ua)
    req.Header.Set("Authorization", r.auth)

    // Debug
    if *r.Flags.Debug {
        // Save a copy of this request for debugging.
        requestDump, err := httputil.DumpRequest(req, true)
        if err != nil {
          fmt.Println(err)
        }
        fmt.Println(string(requestDump))
    }

    // Fire the request
    return http.DefaultClient.Do(req)
}

func (r *Discogs) fetchStr(url string, args map[string]string, Response interface{}) error {
    // Fetch the url
    res, err := r.fetch(url, args)
    if err != nil {
         return err
    }
    defer res.Body.Close()

    // Debug
    if *r.Flags.Debug {
        for k, v := range res.Header {
            fmt.Printf("'%s': %s\n", c(0, 36, k), c(0, 94, v))
        }
    }

    // https://www.discogs.com/developers/#page:home,header:home-rate-limiting
    remainRate, err := strconv.Atoi(res.Header["X-Discogs-Ratelimit-Remaining"][0])
    if err == nil {
        if remainRate < 5 {
            // Wait for a second to avoid hammering the api
            fmt.Println("Throttling...")
            time.Sleep(time.Second)
        }
    }

    // Convert the io.ReadCloser to []byte
    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Body)

    // Unmarshal the result into the Response
    err = json.Unmarshal(buf.Bytes(), Response)
    if err != nil {
         return err
    }

    return nil
}

func (r *Discogs) Search(args map[string]string) (*Search, error) {
    url := "https://api.discogs.com/database/search"
    res := Search{}
    err := r.fetchStr(url, args, &res)
    // Debug
    if *r.Flags.Debug {
        d(res)
    }

    return &res, err
}

func (r *Discogs) GetMasterRelease(id int) (*MasterRelease, error) {
    url := "https://api.discogs.com/masters/" + strconv.Itoa(id)
    res := MasterRelease{}
    err := r.fetchStr(url, args, &res)
    // Debug
    if *r.Flags.Debug {
        d(res)
    }

    return &res, err
}

func (r *Discogs) GetMasterVersions(id int) (*MasterVersions, error) {
    url := "https://api.discogs.com/masters/" + strconv.Itoa(id) + "/versions"
    res := MasterVersions{}

    // Loop endlessly to handle pagination
    for {
        // Create and feed our struct
        mv := MasterVersions{}
        err := r.fetchStr(url, args, &mv)
        if err != nil {
             return nil, err
        }

        // Append the Versions to our response MasterVersions struct
        res.Versions = append(res.Versions, mv.Versions...)

        // Exit the loop if we are in the last page
        if mv.Pagination.Page == mv.Pagination.Pages {
            break
        }

        // Set the url to the next loop
        url = mv.Pagination.Urls.Next
    }
    // Debug
    if *r.Flags.Debug {
        d(res)
    }

    return &res, nil
}

func (r *Discogs) GetRelease(id int) (*Release, error) {
    url := "https://api.discogs.com/releases/" + strconv.Itoa(id)
    res := Release{}
    err := r.fetchStr(url, args, &res)

    //d(res)
    if res.Released == "" {
        res.Released = "0"
    } else {
        // Only Year
        res.Released = res.Released[0:4]
    }

    // Put somethin else if forcing the release id
    if res.Country == "" {
        res.Country = "Unknown"
    }
    // Debug
    if *r.Flags.Debug {
        d(res)
    }

    return &res, err
}

func (r *Discogs) ParseResults(s *Search, q *Query) Results {
    res := Results{
        map[int]*MasterRelease{},
        map[int]*Version{},
    }

    for _, i := range s.Results {
        // Skip if we got unitialized needed variables
        if i.ID == 0 || i.Year == "" || len(i.Label) == 0 || i.Catno == "" || len(i.Format) == 0 {
        //    continue
        }

        // Fetch the versions of this master
        mv, err := r.GetMasterVersions(i.ID)
        check(err)

        // Skip if we got no versions
        if len(mv.Versions) == 0 {
            continue
        }

        // Fetch the concrete Release info from the Master
        mr, err := r.GetMasterRelease(i.ID)
        check(err)

        // Add the MasterRelease to our response map anmd print some info
        res.MasterReleases[mr.ID] = mr
        fmt.Printf("MASTER (%v) Date: %v Genres: %v Styles: %v Items: %v\n", c(0, 97, mr.ID), c(0, 32, mr.Year), c(1, 35, mr.Genres), c(1, 35, mr.Styles), c(1, 35, len(mv.Versions)))

        // Sort the Versions and iterate through them
        sort.Sort(mv.Versions)
        for _, v := range mv.Versions {
            // Skip if we got unitialized needed variables
            if v.ID == 0 || v.Title == "" || v.Released == "" || v.Label == "" || v.Catno == "" || v.Country == "" || v.Format == "" {
                continue
            }

            // Release must have 4 digits at least (year) and be a number
            if len(v.Released) > 3 {
                // Only want the year
                v.Released = v.Released[0:4]
                // Check if it doesn't looks like a number
                if _, err := strconv.Atoi(v.Released); err != nil {
                    continue
                }
            }

            // Default colors for id and release date
            id, released := c(0, 32, v.ID), c(0, 36, v.Released)

            // Use White Bold on the main release
            if v.ID == mr.MainRelease {
                id = c(0, 97, v.ID)
            }

            // Use ia differente year color for newer versions
            if v.Released > strconv.Itoa(mr.Year) {
                released = c(0, 0, v.Released)
            }

            // Build the release string
            str := fmt.Sprintf("\t(%v) %s (%s / %s / %s / %s) [%s]", id, v.Title, released, c(1, 34, v.Label), c(0, 35, v.Catno), c(0, 33, v.Country), v.Format)

            // Match the result to our regexp if we got one
            if *q.Flags.Regexp != "" {
                re := regexp.MustCompile(*q.Flags.Regexp)

                // Skip this version if it doesn't match our regexp
                if !re.MatchString(str) {
                    continue
                }
            }

            // Print the version and add it to the results map
            fmt.Println(str)
            v.MasterID = mr.ID
            res.MasterVersions[v.ID] = v
        }
    }

    return res
}

// Ask user to choose a version
func (r *Discogs) ChooseVersion(res *Results) (int, error) {
    // Leave if got no results
    if len(res.MasterVersions) == 0 {
        return 0, fmt.Errorf("Got no results")
    }

    // Loop forever
    for {
        input := ""
        // Ask user to choose an ID
        fmt.Printf("\nEnter the discogs ID or just <enter> to skip: ")
        fmt.Scanln(&input)

        // Leave if no input was given, only enter
        if input == "" {
            return 0, fmt.Errorf("No input was given.")
        }

        // Check if input is a number
        if index, err := strconv.Atoi(input); err == nil {
            // Check if index is a valid key
            if _, ok := res.MasterVersions[index]; ok {
                return index, nil
            }
        }
    }
}

// Show all the media
func (r *Discogs) PrintTrackList(q *Query) {
    fmt.Println("# Release got different media, we must choose one:")
    for _, track := range q.Release.Tracklist {
        switch t := track.Type; t {
        case "heading":
            fmt.Printf("'%s'\n", c(0, 35, track.Title))
        case "track":
            fmt.Printf("\t%s - %s - (%s)\n", c(0, 36, track.Position), c(0, 36, track.Title), c(0, 36, track.Duration))
        }
    }
}

// Ask user to choose a single media if a release has many
func (r *Discogs) SetMediaName(q *Query) {
    // Loop forever
    for {
        fmt.Printf("\nEnter a string to distinguish this media ('CD1', for example) or just <enter> to skip : ")
        scanner := bufio.NewScanner(os.Stdin)
        if scanner.Scan() {
            q.MediaName = scanner.Text()
            break
        }
    }
}

func main() {
    // Create a dg object for all the run
    dg := NewDiscogs()

    // Parse our commandline flags and arguments
    f := Flags{}
    f.Extra = flag.String("e", "", "Info to add to the directory name like the ripper or source for example.")
    f.Regexp = flag.String("r", "", "Regexp to narrow our queries, 'Vinyl|LP|\"' for example.")
    f.Debug = flag.Bool("d", false, "Debug.")
    f.Id = flag.Int("id", 0, "Force a discogs release id.")
    f.Uid = flag.Int("uid", syscall.Getuid(), "Try to change the process UID to another.")
    flag.Parse()

    // Print the flags info and exit if no dir was provided
    if len(flag.Args()) == 0 {
        pn := filepath.Base(os.Args[0])
        fmt.Printf("Usage:\n\n%s <ARGS> <DIRS>\n\nARGS:\n", pn)
        flag.PrintDefaults()
        fmt.Printf("DIRS:\n  Directories with flac files in the format 'Artist - Album' if possible or named with a Discogs ID.\n")
        os.Exit(0)
    }

    // Print a banner with our flags
    fmt.Printf("Extra: '%v' Regexp: '%v' Id: '%v' UID: '%v' Debug: %s\n",
              c(0, 36, *f.Extra),
              c(0, 36, *f.Regexp),
              c(0, 36, *f.Id),
              c(0, 36, *f.Uid),
              c(0, 36, *f.Debug),
              )
    // Copy the flags to the dg object to have them through all execution
    dg.Flags = f

    // Try to change our UID if choosed
    if *f.Uid != syscall.Getuid() {
        err := syscall.Setuid(*f.Uid)
        if err != nil {
            panic("Failed to change UID")
        }
    }

    // Process every argument (directories)
    for _, dir := range flag.Args() {
        fi, err := os.Stat(dir)
        if err != nil {
            check(err)
            continue
        }
        // Skip if file is not a directory
        if !fi.IsDir() {
            continue
        }

        // The Query struct will hold all the info we need
        q := Query{}
        // Needs to be aware of the flags
        q.Flags = f
        q.Tags = map[string]string{}

        // "Clean returns the shortest path name equivalent to path by purely lexical processing."
        q.OldDir = filepath.Clean(dir)
        q.Query = filepath.Base(q.OldDir)
        q.CleanOldDir = cleanString(q.OldDir)
        q.Query = cleanString(q.Query)
        if strings.HasPrefix(q.Query, "#") {
            continue
        }

        fmt.Printf("Trying '%s'\n", c(1, 33, dir))
        fmt.Printf("Querying discogs with '%s'\n", c(1, 33, q.Query))
        //os.Exit(0)

        id, err := *q.Flags.Id, error(nil)
        if id == 0 {
            // Check if the directory is a discogs Release ID
            id, err = strconv.Atoi(q.Query)
            if err != nil {
                // Split the directory into parts
                v := strings.Split(q.Query, "-")

                // Try to work with a 3 element slice
                if len(v) == 3 {
                    v = append(v[:1], v[2:]...)
                }

                // Complain if we don't havea a proper slice
                if len(v) != 2 {
                    fmt.Println("Search format is not 'Artist - Title'")
                //    continue
                }

                // Builld and launch a master search request
                sArgs:= map[string]string{
                    "type": "master",
                    "per_page": "100",
                    "artist":  strings.TrimSpace(v[0]),
                    "release_title": strings.TrimSpace(v[1]),
                    "query": q.Query,
                }
                s, err := dg.Search(sArgs)

                // Try to make sense of the search results
                res := dg.ParseResults(s, &q)

                // Choose our release manually
                id, err = dg.ChooseVersion(&res)

                // Skip if we got no ID
                if err != nil {
                    fmt.Printf("%s for '%s'\n", err, q.OldDir)
                    continue
                }
            }
        }

        // At this rate we should have a release ID, perform the search
        q.Release, err = dg.GetRelease(id)
        check(err)

        // Get also the master for extra information
        q.Master, err = dg.GetMasterRelease(q.Release.MasterID)
        check(err)

        // Name the release if multiple/
        if q.Release.FormatQuantity != 1 {
            dg.PrintTrackList(&q)
            dg.SetMediaName(&q)
        }

        // Populate the tag map
        buildTags(&q)

        // Create a proper name
        buildName(&q)

        // Check if the directory we would create already exits
        if _, err := os.Stat(q.NewAlbumDir); !os.IsNotExist(err) {
            fmt.Printf("# Directory '%s' already exists!\n", c(1, 31, q.NewAlbumDir))
            fmt.Printf("# The original is '%s'\n", c(1, 31, filepath.Base(q.OldDir)))

            // Ask for disambiguation
            dg.SetMediaName(&q)
            // Compose the name again
            buildName(&q)
        }

        // Print the tags we will add
        fmt.Println("# Common tags to use:")
        d(q.Tags)

        // Retag the flac files and copy the extra files into the new directory
        err = tagAndCopy(&q)
        if err != nil {
            fmt.Println(err)
            continue
        }

        // Fetch the album cover
        getCover(dg, &q)

        fmt.Println("# Moving the original to the #Done folder")
        path, dir := filepath.Split(q.OldDir)
        backup := filepath.Join(path, "#Done", dir)
        backupDir := filepath.Join(path, "#Done")
        // Create all the path, won't fail if it already exists
        os.MkdirAll(backupDir, os.ModePerm)
        err = os.Rename(q.OldDir, backup)
        if err != nil {
            fmt.Println(err)
            continue
        }
    }
}
