install: $(shell find -name "*.go")
	go install

build-all: build-linux build-win

build-linux: $(shell find -name "*.go")
	env GOOS=linux go build -o gopacked

build-win: $(shell find -name "*.go")
	env GOOS=windows go build -o gopacked.exe

debian: build-linux
	mkdir -p build
	mkdir -p package/usr/bin/
	cp gopacked package/usr/bin/
	dpkg-deb --build package gopacked.deb
	mv gopacked.deb build/

linux: build-linux
	mkdir -p build
	tar cvfJ gopacked_linux.tar.xz gopacked LICENSE README.md
	mv gopacked_linux.tar.xz build/

windows: build-win
	mkdir -p build
	zip -9r gopacked_windows.zip gopacked.exe LICENSE README.md
	mv gopacked_windows.zip build/

package: debian linux windows

clean-exes:
	rm -f gopacked gopacked.exe

clean:
	rm -rf build package/usr
