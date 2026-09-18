using Emby.Naming.Common;
using Emby.Server.Implementations.Library.Resolvers.Movies;
using Emby.Server.Implementations.Library.Resolvers.TV;
using Jellyfin.Data.Enums;
using MediaBrowser.Controller;
using MediaBrowser.Controller.Drawing;
using MediaBrowser.Controller.Entities;
using MediaBrowser.Controller.Entities.Movies;
using MediaBrowser.Controller.Entities.TV;
using MediaBrowser.Controller.Library;
using MediaBrowser.Controller.Providers;
using MediaBrowser.Model.IO;
using Microsoft.Extensions.Logging;
using Moq;
using Xunit;

namespace Jellyfin.Server.Implementations.Tests.Library;

public class VideoLibraryResolverTests
{
    private static readonly NamingOptions _namingOptions = new();

    [Fact]
    public void Resolve_MovieCollectionVideoFile_ResolvesMovie()
    {
        var resolver = new MovieResolver(
            Mock.Of<IImageProcessor>(),
            Mock.Of<ILogger<MovieResolver>>(),
            _namingOptions,
            Mock.Of<IDirectoryService>());
        var args = CreateFileArgs(
            CollectionType.movies,
            "/movies/Example Movie (2026)/Example Movie (2026).mkv");

        Assert.IsType<Movie>(resolver.ResolvePath(args));
    }

    [Fact]
    public void Resolve_HomeVideosCollectionVideoFile_ResolvesGenericVideo()
    {
        var resolver = new MovieResolver(
            Mock.Of<IImageProcessor>(),
            Mock.Of<ILogger<MovieResolver>>(),
            _namingOptions,
            Mock.Of<IDirectoryService>());
        var args = CreateFileArgs(
            CollectionType.homevideos,
            "/home-videos/Family Trip 2026.mkv");

        Assert.IsType<Video>(resolver.ResolvePath(args));
    }

    [Fact]
    public void Resolve_TvShowsCollectionEpisodeFile_ResolvesEpisode()
    {
        var resolver = new EpisodeResolver(
            Mock.Of<ILogger<EpisodeResolver>>(),
            _namingOptions,
            Mock.Of<IDirectoryService>());
        var args = CreateFileArgs(
            CollectionType.tvshows,
            "/tv/Example Show/Season 01/Example Show S01E01.mkv");

        Assert.IsType<Episode>(resolver.ResolvePath(args));
    }

    private static ItemResolveArgs CreateFileArgs(CollectionType collectionType, string path)
    {
        return new ItemResolveArgs(
            Mock.Of<IServerApplicationPaths>(),
            Mock.Of<ILibraryManager>())
        {
            Parent = new Folder { Name = "Library" },
            CollectionType = collectionType,
            FileInfo = new FileSystemMetadata
            {
                FullName = path,
                IsDirectory = false
            }
        };
    }
}
