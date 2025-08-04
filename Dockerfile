# Gunakan image Go resmi dari Docker Hub
FROM golang:1.24.4-alpine

# Set working directory di dalam container
WORKDIR /app

# Copy file go.mod dan go.sum terlebih dahulu untuk memanfaatkan cache
COPY go.mod go.sum ./

# Install dependencies Go
RUN go mod tidy

# Copy seluruh kode aplikasi ke dalam container
COPY . .

# Expose port yang akan digunakan oleh aplikasi
EXPOSE 8080

# Perintah untuk menjalankan aplikasi Go
CMD ["go", "run", "main.go"]