mkdir test_walker
touch test_walker/file1
touch test_walker/file2
go run record.go test_walker
rm test_walker/file1
go run record.go test_walker
echo "new" > test_walker/file2
go run record.go test_walker
