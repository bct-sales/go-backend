#!/usr/bin/env ruby


VERSION_FILE = 'version/version.txt'


version_string = IO.read(VERSION_FILE).strip
major, minor, patch = version_string.split('.').map(&:to_i)


case ARGV[0]
when 'major'
  major += 1
  minor = 0
  patch = 0

when 'minor'
  minor += 1
  patch = 0

when 'patch'
  patch += 1

else
  abort "Expected major, minor or patch as argument"
end

new_version_string = "#{major}.#{minor}.#{patch}"
IO.write(VERSION_FILE, new_version_string)
`git add #{VERSION_FILE}`
`git commit -m "#{new_version_string}`
`git tag v#{new_version_string}`
puts "Updated to #{new_version_string}"