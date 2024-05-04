# DApps API [![Go](https://github.com/jaganathanb/dapps-api/actions/workflows/go.yml/badge.svg)](https://github.com/jaganathanb/dapps-api/actions/workflows/go.yml)

#Build the API

```
cd src\cmd

go-winres make --product-version=git-tag --file-version=git-tag

cd ..
cd ..

#open bash terminal and run the below cmd

bash build.sh

```

#Create tag & push to remote

git tag <version number> # ex. git tag v1.4.1.0

git push origin v1.4.0.0 # pushing the tag to remote