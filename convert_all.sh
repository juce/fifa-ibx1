#!/bin/bash -ue
mkdir -p xmls
find dat -name '*.DAT' -type f | while read x; do
    y=$(echo $x | sed "s/.DAT$/.xml/g")
    y=xmls/$(basename $y)
    echo "converting $x --> $y"
    ./reader.py $x 2>>debug.log 1>$y
done
echo converted $(ls xmls/*.xml | wc -l) files
