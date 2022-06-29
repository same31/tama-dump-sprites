# tama-dump-sprites

Command line utility to extract sprites from a Tamagotchi dump

```bash
# Build
$ go build -o tama-dump-sprites src/main.go

# Display help
$ ./tama-dump-sprites -h

# Extract sprites from a dump file into the current directory
$ ./tama-dump-sprites -i dump.bin

# Extract sprites from a dump file into specified directory
$ ./tama-dump-sprites -i dump.bin -o ~/sprites/

# Extract raw sprites with a green background instead of a transparent one
$ ./tama-dump-sprites -i dump.bin -r
```

## Credits

I translated in Golang the image decoding part found in this repo: https://github.com/kaero/t-on 