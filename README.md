# blockhash

A Go port of https://github.com/commonsmachinery/blockhash-python.

## Usage
```bash
go build
./blockhash path_to_image.jpg
```
This will output a 64 character hexadecimal hash representing the image.


## Known issues
The hashes produced are not the same as other versions because the Go image library reads different pixel values (sometimes, but consistently).

## License
MIT
