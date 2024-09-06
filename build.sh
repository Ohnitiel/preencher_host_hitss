echo "Compiling for MacOS"
GOOS=darwin go build -o build/preencher_host_macos
echo "Done MacOS"

echo "Compiling for Windows"
GOOS=windows go build -o build/preencher_host_windows.exe
echo "Done for Windows"

echo "Compiling for Linux"
GOOS=linux go build -o build/preencher_host_linux
echo "Done for Linux"

