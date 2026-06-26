rm -rf ../builds

mkdir ../builds
mkdir ../builds/linux
mkdir ../builds/windows

go build -o ../builds/linux/server ./cmd/server
go build -o ../builds/linux/setup ./cmd/setup

GOOS=windows GOARCH=amd64 go build -o ../builds/windows/server.exe ./cmd/server/
GOOS=windows GOARCH=amd64 go build -o ../builds/windows/setup.exe ./cmd/setup/

cp -r ../db ../builds/linux/db
cp ../README.md ../builds/linux/README.md
cp ../LICENSE ../builds/linux/LICENSE

cp -r ../db ../builds/windows/db
cp ../README.md ../builds/windows/README.md
cp ../LICENSE ../builds/windows/LICENSE

zip -r ../builds/linux-bin.zip ../builds/linux/
zip -r ../builds/windows-bin.zip ../builds/windows/

cp -r ../src ../builds/linux/src
cp -r ../src ../builds/windows/src

zip -r ../builds/linux.zip ../builds/linux/
zip -r ../builds/windows.zip ../builds/windows/
