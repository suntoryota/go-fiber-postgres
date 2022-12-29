package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-fiber-postgres/models"
	"go-fiber-postgres/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}
type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	book := Book{}
	err := context.BodyParser(&book)
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"Message": "failed to create book"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been added",
	})
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "failed get books"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "books fetched successfully",
		"data":    bookModels,
	})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cant be empty",
		})
		return nil
	}
	err := r.DB.Delete(bookModel, id)
	if err != nil {
		context.Status(http.StatusOK).JSON(&fiber.Map{
			"message": "books delete successfully",
		})
		return nil
	}
	return nil
}

func (r *Repository) GetBookById(context *fiber.Ctx) error {
	id := context.Params("id")
	bookModel := &models.Books{}

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cant be empty",
		})
		return nil
	}
	fmt.Println("The ID is", id)

	err := r.DB.Where("id = ?", id).First(bookModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "cant get book",
		})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book id fetched successfully",
		"data":    bookModel,
	})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookById)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("Failed to connect database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("cant migrate db")
	}
	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")
}
