#!/bin/bash

file="version.txt" #the file where you keep your string name

current_version=$(cat "$file")        #the output of 'cat $file' is assigned to the $name variable

go build -o test-goreleaser

new_version=$(echo $current_version | awk -F. -v OFS=. 'NF==1{print ++$NF}; NF>1{if(length($NF+1)>length($NF))$(NF-1)++; $NF=sprintf("%0*d", length($NF), ($NF+1)%(10^length($NF))); print}')

echo "$new_version" > $file
new_image="releaser:$new_version"


echo "Deploying $new_image"

docker build -t $new_image .
docker tag $new_image registry.geff.ws/test-goreleaser/$new_image
docker push registry.geff.ws/test-goreleaser/$new_image

echo "Nova versão: $new_version"