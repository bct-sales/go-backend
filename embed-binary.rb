input = ARGV[0]

File.open(input, 'rb') do |f|
  $data = f.read.bytes
end

bytes = $data.map(&:to_s).join(", ")
puts("var imageData = []byte{#{bytes}}")