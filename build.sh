# !/bin/bash
DIRS=$(find . -type d | grep "./cmd//*" | cut -f 3 -d /)
for DIR in $DIRS
do 
    go build -o build/bin/$DIR cmd/$DIR/main.go
done