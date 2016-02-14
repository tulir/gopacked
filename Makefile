install: $(shell find -name "*.go")
	go install

linux: $(shell find -name "*.go")
	go build
	zip -9r gopacked_linux.zip gopacked LICENSE GABS_LICENSE PFLAG_LICENSE README.md
	rm -f gopacked

windows: $(shell find -name "*.go")
	env GOOS=windows go build
	zip -9r gopacked_windows.zip gopacked.exe LICENSE GABS_LICENSE PFLAG_LICENSE README.md
	rm -f gopacked.exe

clean:
	rm -f gopacked gopacked.exe gopacked.zip
