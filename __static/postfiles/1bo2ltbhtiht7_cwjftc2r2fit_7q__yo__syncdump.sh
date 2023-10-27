if [ -z "$1" ]; then
    echo -e "empty 1st arg"
    exit 1
fi
if [ -z "$2" ]; then
    echo -e "empty 2nd arg"
    exit 1
fi

echo ===CHMOD===
echo
sudo chmod -R 0777 "$2"
echo
echo ===SYNC===
echo
dumpsync -max "$1" -out "$2"
echo
echo ===UMOUNT===
echo
udevil umount -f "$2"
