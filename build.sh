if command -v wgo &> /dev/null; then
  GO=wgo
else
  GO=go
fi

$GO mod tidy

$GO build -o main ./src
