ifeq ($(OS),Windows_NT)
    NEW_TERM = start pwsh -NoExit -Command
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Darwin)
        NEW_TERM = osascript -e 'tell application "Terminal" to do script'
    else
        # Assume Linux with gnome-terminal available
        NEW_TERM = gnome-terminal -- 
    endif
endif

dev:
	$(NEW_TERM) "go run . :5050"
