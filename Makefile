ifeq ($(OS), Windows_NT)
	EXECUTABLE=$(MODULE_NAME).exe
	BUILD_FLAGS=-ldflags '-extldflags "-static" -H=windowsgui' .
else
	EXECUTABLE=$(MODULE_NAME)
	BUILD_FLAGS=.
endif

static-build:
	go build $(BUILD_FLAGS)