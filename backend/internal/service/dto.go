package service

// DTOs mirror specs/api/contract.md exactly. Optional values are pointers
// without omitempty, so they are serialised as null and never omitted.

// PhotoDTO is the photo object (contract §2.5).
type PhotoDTO struct {
	URL      string  `json:"url"`
	ThumbURL *string `json:"thumbUrl"`
	Alt      *string `json:"alt"`
	Caption  *string `json:"caption"`
	Width    *int    `json:"width"`
	Height   *int    `json:"height"`
}

// LinkDTO is a contact link.
type LinkDTO struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

// MainDTO is the GET /api/v1/main response (contract §2.2).
type MainDTO struct {
	FullName  string    `json:"fullName"`
	Headline  string    `json:"headline"`
	Location  string    `json:"location"`
	Currently string    `json:"currently"`
	Bio       []string  `json:"bio"`
	Photo     PhotoDTO  `json:"photo"`
	Interests []string  `json:"interests"`
	Links     []LinkDTO `json:"links"`
}

// ExperienceDTO is one job on the CV.
type ExperienceDTO struct {
	ID          string   `json:"id"`
	Position    string   `json:"position"`
	Company     string   `json:"company"`
	CompanyURL  *string  `json:"companyUrl"`
	Location    string   `json:"location"`
	StartDate   string   `json:"startDate"`
	EndDate     *string  `json:"endDate"`
	Description *string  `json:"description"`
	Highlights  []string `json:"highlights"`
}

// EducationDTO is one education entry on the CV.
type EducationDTO struct {
	ID          string  `json:"id"`
	Institution string  `json:"institution"`
	Degree      string  `json:"degree"`
	Field       string  `json:"field"`
	StartDate   string  `json:"startDate"`
	EndDate     *string `json:"endDate"`
	Description *string `json:"description"`
}

// SkillGroupDTO is a skill category.
type SkillGroupDTO struct {
	Category string   `json:"category"`
	Items    []string `json:"items"`
}

// LanguageDTO is a spoken language.
type LanguageDTO struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

// CVDTO is the GET /api/v1/cv response (contract §2.3).
type CVDTO struct {
	Summary    string          `json:"summary"`
	Experience []ExperienceDTO `json:"experience"`
	Education  []EducationDTO  `json:"education"`
	Skills     []SkillGroupDTO `json:"skills"`
	Languages  []LanguageDTO   `json:"languages"`
}
