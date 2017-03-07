FROM alpine
ADD gopath/bin/cloud-native-app /cloud-native-app
ENTRYPOINT ["/cloud-native-app"]
