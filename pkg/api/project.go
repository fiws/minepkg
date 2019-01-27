package api

import "context"

// Project returns a Project object without fetching it from the API
func (m *MinepkgAPI) Project(name string) *Project {
	return &Project{
		c:    m,
		Name: name,
	}
}

// GetProject gets a single project
func (m *MinepkgAPI) GetProject(ctx context.Context, name string) (*Project, error) {
	res, err := m.get(ctx, baseAPI+"/projects/"+name)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(res); err != nil {
		return nil, err
	}

	project := Project{c: m}
	if err := parseJSON(res, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// CreateProject creates a new project
func (m *MinepkgAPI) CreateProject(p *Project) (*Project, error) {
	res, err := m.postJSON(context.TODO(), baseAPI+"/projects", p)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(res); err != nil {
		return nil, err
	}

	project := Project{}
	if err := parseJSON(res, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// GetReleases gets a all available releases for this project
func (p *Project) GetReleases(ctx context.Context) ([]*Release, error) {
	res, err := p.c.get(ctx, baseAPI+"/projects/"+p.Name+"/releases")
	if err != nil {
		return nil, err
	}
	if err := checkResponse(res); err != nil {
		return nil, err
	}

	releases := make([]*Release, 0, 20)
	if err := parseJSON(res, &releases); err != nil {
		return nil, err
	}
	for _, r := range releases {
		r.decorate(p.c) // sets the private client field
	}

	return releases, nil
}
