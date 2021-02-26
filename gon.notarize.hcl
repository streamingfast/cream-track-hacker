source = ["./dist/tracker.zip"]
bundle_id = "io.streamingfast.cream-track-hacker.cmd"

apple_id {
  # The username when not defined is picked automatically from env var AC_USERNAME
  # The password when not defined is picked automatically from env var AC_PASSWORD
}

sign {
  application_identity = "Developer ID Application: dfuse Platform Inc. (ZG686LRL8C)"
}

notarize {
  path = "./dist/tracker.zip"
  bundle_id = "io.streamingfast.cream-track-hacker.cmd"
}
