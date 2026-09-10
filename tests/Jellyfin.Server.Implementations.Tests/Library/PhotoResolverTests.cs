using System;
using System.Collections.Generic;
using Emby.Naming.Common;
using Emby.Server.Implementations.Library.Resolvers;
using Jellyfin.Data.Enums;
using MediaBrowser.Controller;
using MediaBrowser.Controller.Drawing;
using MediaBrowser.Controller.Entities;
using MediaBrowser.Controller.Library;
using MediaBrowser.Controller.Providers;
using MediaBrowser.Model.Configuration;
using MediaBrowser.Model.IO;
using Moq;
using Xunit;

namespace Jellyfin.Server.Implementations.Tests.Library;

public class PhotoResolverTests
{
    private static readonly NamingOptions _namingOptions = new();

    [Fact]
    public void PhotoResolver_HomeVideosWithPhotosEnabled_DoesNotResolvePhoto()
    {
        var imageProcessor = CreateImageProcessor();
        var directoryService = new Mock<IDirectoryService>();
        directoryService
            .Setup(service => service.GetFiles(It.IsAny<string>()))
            .Returns(new List<FileSystemMetadata>());
        var resolver = new PhotoResolver(imageProcessor.Object, _namingOptions, directoryService.Object);
        var args = CreateFileArgs(CollectionType.homevideos, enablePhotos: true);

        Assert.Null(resolver.ResolvePath(args));
    }

    [Fact]
    public void PhotoResolver_PhotosCollection_ResolvesPhoto()
    {
        var imageProcessor = CreateImageProcessor();
        var directoryService = new Mock<IDirectoryService>();
        directoryService
            .Setup(service => service.GetFiles(It.IsAny<string>()))
            .Returns(new List<FileSystemMetadata>());
        var resolver = new PhotoResolver(imageProcessor.Object, _namingOptions, directoryService.Object);
        var args = CreateFileArgs(CollectionType.photos, enablePhotos: false);

        Assert.IsType<Photo>(resolver.ResolvePath(args));
    }

    [Fact]
    public void PhotoAlbumResolver_HomeVideosWithPhotosEnabled_DoesNotResolvePhotoAlbum()
    {
        var imageProcessor = CreateImageProcessor();
        var resolver = new PhotoAlbumResolver(imageProcessor.Object, _namingOptions);
        var args = CreateDirectoryArgs(CollectionType.homevideos, enablePhotos: true);

        Assert.Null(resolver.ResolvePath(args));
    }

    [Fact]
    public void PhotoAlbumResolver_PhotosCollection_ResolvesPhotoAlbum()
    {
        var imageProcessor = CreateImageProcessor();
        var resolver = new PhotoAlbumResolver(imageProcessor.Object, _namingOptions);
        var args = CreateDirectoryArgs(CollectionType.photos, enablePhotos: false);

        Assert.IsType<PhotoAlbum>(resolver.ResolvePath(args));
    }

    private static Mock<IImageProcessor> CreateImageProcessor()
    {
        var imageProcessor = new Mock<IImageProcessor>();
        imageProcessor
            .SetupGet(processor => processor.SupportedInputFormats)
            .Returns(new[] { "jpg" });
        return imageProcessor;
    }

    private static ItemResolveArgs CreateFileArgs(CollectionType collectionType, bool enablePhotos)
    {
        return new ItemResolveArgs(Mock.Of<IServerApplicationPaths>(), Mock.Of<ILibraryManager>())
        {
            CollectionType = collectionType,
            LibraryOptions = new LibraryOptions { EnablePhotos = enablePhotos },
            FileInfo = new FileSystemMetadata
            {
                FullName = "/library/photo.jpg",
                Name = "photo.jpg",
                IsDirectory = false
            },
            FileSystemChildren = Array.Empty<FileSystemMetadata>()
        };
    }

    private static ItemResolveArgs CreateDirectoryArgs(CollectionType collectionType, bool enablePhotos)
    {
        return new ItemResolveArgs(Mock.Of<IServerApplicationPaths>(), Mock.Of<ILibraryManager>())
        {
            CollectionType = collectionType,
            LibraryOptions = new LibraryOptions { EnablePhotos = enablePhotos },
            FileInfo = new FileSystemMetadata
            {
                FullName = "/library/album",
                Name = "album",
                IsDirectory = true
            },
            FileSystemChildren =
            [
                new FileSystemMetadata
                {
                    FullName = "/library/album/photo.jpg",
                    Name = "photo.jpg",
                    IsDirectory = false
                }
            ]
        };
    }
}
