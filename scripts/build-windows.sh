#! /bin/bash

echo -e "Start running the script..."
cd ../
export CGO_ENABLED=0

echo -e "Start building the frontend..."
(cd frontend && npm run build)
test -f frontend/dist/index.html

echo -e "Start building the app for windows platform..."
wails build --clean -s --platform windows/amd64

echo -e "End running the script!"
