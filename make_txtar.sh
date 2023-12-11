rm -rf /tmp/outbow
rm -f /tmp/outbow.tar
rm -f /tmp/filelist.txt

{
    git ls-files |
        grep -v README.org |
        grep -v make_txtar.sh |
        grep -v urls.db |
        grep -v go.sum |
        grep -v Makefile |
        grep -v gopro.scpt.tmpl |
        grep -v gopro0252.scpt |
        grep -v go.mod
} | tee /tmp/filelist.txt

tar -cf /tmp/outbow.tar -T /tmp/filelist.txt
mkdir -p /tmp/outbow
tar xf /tmp/outbow.tar -C /tmp/outbow
rg --files /tmp/outbow
txtar-c /tmp/outbow | pbcopy
