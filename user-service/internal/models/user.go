package models

import "time"

// User - базовая модель пользователя
type User struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"not null" json:"role"` // "creator" или "venue"
	AvatarID     *int      `json:"avatar_id,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Связи
	Avatar  *Image   `gorm:"foreignKey:AvatarID" json:"avatar,omitempty"`
	Creator *Creator `gorm:"foreignKey:UserID" json:"creator,omitempty"`
	Venue   *Venue   `gorm:"foreignKey:UserID" json:"venue,omitempty"`
}

// Creator - профиль создателя мероприятий
type Creator struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int       `gorm:"not null" json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	PhotoID     *int      `json:"photo_id,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	WorkEmail   string    `json:"work_email,omitempty"`
	TgPersonal  string    `gorm:"column:tg_personal_link" json:"tg_personal_link,omitempty"`
	TgChannel   string    `gorm:"column:tg_channel_link" json:"tg_channel_link,omitempty"`
	VkLink      string    `json:"vk_link,omitempty"`
	TiktokLink  string    `json:"tiktok_link,omitempty"`
	YoutubeLink string    `json:"youtube_link,omitempty"`
	DzenLink    string    `json:"dzen_link,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Связи
	User  User   `gorm:"foreignKey:UserID" json:"-"`
	Photo *Image `gorm:"foreignKey:PhotoID" json:"photo,omitempty"`
}

// Venue - профиль площадки
type Venue struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int       `gorm:"not null" json:"user_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Address      string    `json:"address,omitempty"`
	OpeningHours string    `json:"opening_hours,omitempty"`
	Capacity     int       `json:"capacity,omitempty"`
	LogoID       *int      `json:"logo_id,omitempty"`
	CoverPhotoID *int      `json:"cover_photo_id,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	WorkEmail    string    `json:"work_email,omitempty"`
	TgPersonal   string    `gorm:"column:tg_personal_link" json:"tg_personal_link,omitempty"`
	TgChannel    string    `gorm:"column:tg_channel_link" json:"tg_channel_link,omitempty"`
	VkLink       string    `json:"vk_link,omitempty"`
	TiktokLink   string    `json:"tiktok_link,omitempty"`
	YoutubeLink  string    `json:"youtube_link,omitempty"`
	DzenLink     string    `json:"dzen_link,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Связи
	User        User          `gorm:"foreignKey:UserID" json:"-"`
	Logo        *Image        `gorm:"foreignKey:LogoID" json:"logo,omitempty"`
	CoverPhoto  *Image        `gorm:"foreignKey:CoverPhotoID" json:"cover_photo,omitempty"`
	Photos      []VenuePhoto  `gorm:"foreignKey:VenueID" json:"photos,omitempty"`
	Categories  []int         `gorm:"-" json:"category_ids,omitempty"` // Только для передачи данных
}

// VenuePhoto - дополнительные фото площадки
type VenuePhoto struct {
	ID      int `gorm:"primaryKey;autoIncrement" json:"id"`
	VenueID int `gorm:"not null" json:"venue_id"`
	ImageID int `gorm:"not null" json:"image_id"`

	Image Image `gorm:"foreignKey:ImageID" json:"image"`
}

// VenueCategory - связь площадки с категориями
type VenueCategory struct {
	VenueID    int `gorm:"primaryKey" json:"venue_id"`
	CategoryID int `gorm:"primaryKey" json:"category_id"`
}

// Image - метаданные изображений из MinIO
type Image struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	FileName   string    `gorm:"not null" json:"file_name"`
	FilePath   string    `gorm:"not null" json:"file_path"`
	FileType   string    `json:"file_type,omitempty"`
	ImageType  string    `gorm:"not null" json:"image_type"` // avatar, venue-logo, venue-cover, venue-photo, event-cover
	BucketName string    `gorm:"not null" json:"bucket_name"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName overrides
func (User) TableName() string          { return "users" }
func (Creator) TableName() string       { return "creators" }
func (Venue) TableName() string         { return "venues" }
func (VenuePhoto) TableName() string    { return "venue_photos" }
func (VenueCategory) TableName() string { return "venue_categories" }
func (Image) TableName() string         { return "images" }
