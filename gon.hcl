# The path follows a pattern
# ./dist/BUILD-ID_TARGET/BINARY-NAME
source = ["./dist/tracker-osx_darwin_amd64/tracker"]
bundle_id = "io.streamingfast.project.id"

apple_id {
  username = "@env:AC_EMAIL"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: dfuse Platform Inc. (ZG686LRL8C)"
}