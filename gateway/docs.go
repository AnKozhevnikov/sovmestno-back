package main

// @title           Совместно API Gateway
// @version         1.0
// @description     API Gateway для проекта "Совместно". Проксирует запросы к микросервисам и обеспечивает аутентификацию.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@sovmestno.ru

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name health
// @tag.description Health check endpoint

// @tag.name user-service
// @tag.description Проксирование запросов к User Service
