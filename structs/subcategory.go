package structs

type Subcategory struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryId  int64  `json:"category_id"`
	CreatedAt   string `json:"created_at"`
	ImageUrl    string `json:"image_url"`
}
