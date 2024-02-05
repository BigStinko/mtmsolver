package tmdbapi

type MovieResource struct {
	Title string `json:"title"`
	Id    int	 `json:"id"`
}

type ActorResource struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

type MovieQueryResult struct {
	Results      []MovieResource `json:"results"`
	Page         int             `json:"page"`
	TotalPages   int             `json:"total_pages"`
	TotalResults int             `json:"total_results"`
}

type ActorQueryResult struct {
	Results      []ActorResource `json:"results"`
	Page         int             `json:"page"`
	TotalPages   int             `json:"total_pages"`
	TotalResults int             `json:"total_results"`
}

type Credits struct {
	Cast []struct{
		Id        int    `json:"id"`
		Character string `json:"order"`
	} `json:"cast"`
}

type MovieCredits struct {
	Cast []struct {
		Id	      int    `json:"id"`
		Character string `json:"order"`
	} `json:"cast"`
}
