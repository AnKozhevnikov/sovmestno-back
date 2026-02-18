package main

// @title           Совместно API - User Service
// @version         1.0
// @description     API для управления пользователями, профилями создателей и площадок
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@sovmestno.ru

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8081
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name auth
// @tag.description Аутентификация и авторизация

// @tag.name creators
// @tag.description Управление профилями создателей мероприятий

// @tag.name venues
// @tag.description Управление профилями площадок
