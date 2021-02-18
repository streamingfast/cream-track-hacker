# The path follows a pattern
# ./dist/BUILD-ID_TARGET/BINARY-NAME
source = ["./dist/tracker-osx_darwin_amd64/tracker"]
bundle_id = "io.streamingfast.cream-track-hacker.cmd"

apple_id {
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: dfuse Platform Inc. (ZG686LRL8C)"
}

dmg {
    output_path = "./release/darwin/tracker.dmg"
    volume_name = "tracker"
}

zip {
    output_path = "./release/darwin/tracker.zip"
}
