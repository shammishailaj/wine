package wine

type Controller interface {
	RoutePath() string
	RouteMap() map[string]Handler
}

var _ Controller = (*EmptyController)(nil)

type EmptyController struct {
}

func (ec *EmptyController) RoutePath() string {
	return "/"
}

func (ec *EmptyController) RouteMap() map[string]Handler {
	//return map[string]Handler {
	//	"GET|POST hello":nil,
	//	"Any /world":nil,
	//}
	return nil
}
