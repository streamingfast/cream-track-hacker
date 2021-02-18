# The path follows a pattern
# ./dist/BUILD-ID_TARGET/BINARY-NAME
source = ["./dist/tracker-osx_darwin_amd64/tracker"]
bundle_id = "io.streamingfast.cream-track-hacker.cmd"

apple_id {
  # The username when not defined is picked automatically from env var AC_USERNAME
  # The password when not defined is picked automatically from env var AC_PASSWORD
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
