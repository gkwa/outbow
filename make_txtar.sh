set -x

rm -rf /tmp/outbow
rm -f /tmp/outbow.tar
rm -f /tmp/filelist.txt

{
    rg --files ./ |
        grep -v go.sum |
        grep -v go.mod |
        grep -v README.org |
        grep -v gopro000001.scpt |
        grep -v ./outbow |
        grep -v gopro.scpt |
        grep -v gopro.tmpl |
        grep -v make_txtar.sh |
        grep -v gopro000005.scpt |
        grep -v storage_file.go |
        grep -v Makefile |
        grep -v storage_db.go
} | tee /tmp/filelist.txt

tar -cf /tmp/outbow.tar -T /tmp/filelist.txt
mkdir -p /tmp/outbow
tar xf /tmp/outbow.tar -C /tmp/outbow
rg --files /tmp/outbow
txtar-c /tmp/outbow | pbcopy
