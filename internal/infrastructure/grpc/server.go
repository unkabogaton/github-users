package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	entities "github.com/unkabogaton/github-users/internal/domain/entities"
	derr "github.com/unkabogaton/github-users/internal/domain/errors"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"
	gen "github.com/unkabogaton/github-users/internal/infrastructure/grpc/gen"
)

type Server struct {
	gen.UnimplementedUserServiceServer
	userService interfaces.UserService
}

func NewServer(userService interfaces.UserService) *Server {
	return &Server{userService: userService}
}

func (server *Server) ListenAndServe(address string) error {
	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	grpcServerInstance := grpc.NewServer()
	gen.RegisterUserServiceServer(grpcServerInstance, server)
	reflection.Register(grpcServerInstance)
	return grpcServerInstance.Serve(tcpListener)
}

func mapUserEntityToProto(userEntity *entities.User) *gen.User {
	return &gen.User{
		Id:           int64(userEntity.ID),
		Login:        userEntity.Login,
		NodeId:       userEntity.NodeID,
		AvatarUrl:    userEntity.AvatarURL,
		Url:          userEntity.URL,
		HtmlUrl:      userEntity.HTMLURL,
		Type:         userEntity.Type,
		UserViewType: userEntity.UserViewType,
		SiteAdmin:    userEntity.SiteAdmin,
	}
}

func (server *Server) ListUsers(
	ctx context.Context,
	req *gen.ListUsersRequest,
) (*gen.UserList, error) {

	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 10
	}
	page := int(req.GetPage())
	if page == 0 {
		page = 1
	}
	orderBy := req.GetOrderBy()
	if orderBy == "" {
		orderBy = "id"
	}
	orderDirection := req.GetOrderDirection()
	if orderDirection == "" {
		orderDirection = "asc"
	}

	listOptions := interfaces.ListOptions{
		Limit:          limit,
		Page:           page,
		OrderBy:        orderBy,
		OrderDirection: orderDirection,
	}

	userEntities, err := server.userService.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	protoUsers := make([]*gen.User, 0, len(userEntities))
	for i := range userEntities {
		protoUsers = append(protoUsers, mapUserEntityToProto(&userEntities[i]))
	}

	return &gen.UserList{Users: protoUsers}, nil
}

func (server *Server) GetUser(ctx context.Context, request *gen.GetUserRequest) (*gen.User, error) {
	username := request.GetUsername()
	userEntity, err := server.userService.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	return mapUserEntityToProto(userEntity), nil
}

func (server *Server) UpdateUser(ctx context.Context, request *gen.UpdateUserRequest) (*gen.User, error) {
	updateRequest := interfaces.UpdateUserRequest{
		Login:        request.GetLogin(),
		NodeID:       request.GetNodeId(),
		AvatarURL:    request.GetAvatarUrl(),
		URL:          request.GetUrl(),
		HTMLURL:      request.GetHtmlUrl(),
		Type:         request.GetType(),
		UserViewType: request.GetUserViewType(),
		SiteAdmin:    request.GetSiteAdmin(),
	}

	updatedUser, err := server.userService.Update(ctx, request.GetUsername(), updateRequest)
	if err != nil {
		return nil, err
	}

	return mapUserEntityToProto(updatedUser), nil
}

func (server *Server) DeleteUser(ctx context.Context, request *gen.DeleteUserRequest) (*gen.DeleteUserResponse, error) {
	username := request.GetUsername()
	if username == "" {
		return nil, derr.New(derr.ErrorCodeValidation, "username is required")
	}

	if err := server.userService.Delete(ctx, username); err != nil {
		return nil, err
	}

	return &gen.DeleteUserResponse{Username: username}, nil
}
