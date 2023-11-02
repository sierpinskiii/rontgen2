# Use the latest Ubuntu as the base image
FROM ubuntu:latest

EXPOSE 8080

# Update and install necessary dependencies
RUN apt-get update && \
    apt-get install -y git golang-go imagemagick && \
    rm -rf /var/lib/apt/lists/*

# Clone the repository
# RUN git clone https://github.com/sierpinskiii/rontgen2.git

# Change working directory to the cloned repo
WORKDIR /rontgen2

COPY . .

# Build the Go application
RUN go build

# Remove PDF limitations from ImageMagick's policy.xml file
RUN sed -i '/PDF/s/policy/none/' /etc/ImageMagick-6/policy.xml

# Set the entrypoint to run the built application
ENTRYPOINT ["./rontgen2"]
