# discgos
Tag flac albums using the discogs.com API. Its mainly intended to be used to tag vinyl rips in a interactive way.  

## Notes
* It will query the api with the album's directory name, the recommended format is 'Artist - Album'.  
* If there is any result you will have to choose from a list and the album will be tagged in a new dir keeping the original.  
* It will also fetch the cover from the release.  
* If the album directory is named using a discogs release (2287669 for example) it will force tag the album.  
* If the discogs track number matches the directory files it will use the discogs track names, if not, the flac files track names will remain untouched.  
* A regexp can be used to get precise queries, 'Vinyl|LP|\d"' for example will only show vinyl releases. (-r)  
* A string can be use to add info like the ripper or source. (-e)  

## Usage
```
Usage:

discgos <ARGS> <DIRS>

ARGS:
  -d    Debug.
  -e string
        Info to add to the directory name like the ripper or source for example.
  -id int
        Force a discogs release id.
  -r string
        Regexp to narrow our queries, 'Vinyl|LP|"' for example.
  -uid int
        Try to change the process UID to another.
DIRS:
  Directories with flac files in the format 'Artist - Album' if possible or named with a Discogs ID.
```

## Example
```
# go run discgos.go -e oprietop@2015 -r 'Vinyl|LP|\d"' Sleep\ -\ Dopesmoker
```

## Thanks
https://www.discogs.com/
