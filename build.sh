bin_name='DApps.WebApi'  # Change the binary name as desired
cd src/cmd

archs=("386" "amd64")
GOOS="windows"

# Look for signs of trouble in each log
for i in ${!archs[@]};
do
arch=${archs[$i]}
echo "Building for ${arch}..."

CGO_ENABLED=1 GOOS="${GOOS}" GOARCH=${arch} \
         go build -o "../../out/${bin_name}-${arch}.exe"
done

rm -rf *.syso