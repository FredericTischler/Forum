# Utiliser l'image officielle de Go (remplacez '1.21' par '1.23' si disponible)
FROM golang:1.23

RUN apt-get update && apt-get install -y gcc libc6-dev

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de dépendances si vous utilisez des modules Go
ENV CGO_ENABLED=1
COPY go.mod go.sum ./

# Télécharger les dépendances
RUN go mod download

# Copier le reste du code source
COPY . .

# Construire l'application
RUN go build -o forum ./server/main.go

# Exposer le port sur lequel votre application écoute (remplacez '8080' par votre port)
EXPOSE 8080

# Démarrer l'application
CMD ["./forum"]