#!/usr/bin/env ruby

# https://github.community/t5/GitHub-Actions/Support-for-YAML-anchors/m-p/42517/highlight/true#M5024

require "yaml"
require "json"

yaml = YAML.load_file(File.expand_path("test.yml", __dir__))

File.write(File.expand_path("../workflows/test_generated.yml", __dir__), YAML.load(yaml.to_json).to_yaml())
