# Estágio 1: Build - Usa uma imagem Go para compilar a aplicação
FROM golang:1.24-alpine AS builder

# Instala o pacote de dados de fuso horário (timezone)
RUN apk add --no-cache tzdata

WORKDIR /app

# Copia os arquivos de dependências e baixa os módulos
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código-fonte
COPY . .

# Compila a aplicação, criando um executável estático
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Estágio 2: Produção - Usa uma imagem mínima para rodar a aplicação
FROM alpine:latest

WORKDIR /root/

# Copia o executável compilado do estágio de build
COPY --from=builder /app/main .

# Copia os dados de fuso horário do estágio de build para a imagem final
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Expõe a porta que a aplicação vai rodar
EXPOSE 3000

# Comando para iniciar o servidor
CMD ["./main"]
