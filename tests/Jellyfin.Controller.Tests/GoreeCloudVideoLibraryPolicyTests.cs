using System;
using MediaBrowser.Controller.Library;
using MediaBrowser.Model.Entities;
using Xunit;

namespace Jellyfin.Controller.Tests
{
    public class GoreeCloudVideoLibraryPolicyTests
    {
        [Fact]
        public void SupportedCollectionTypes_ContainsOnlyVideoLibraryTypes()
        {
            Assert.Equal(
                new[]
                {
                    CollectionTypeOptions.movies,
                    CollectionTypeOptions.tvshows,
                    CollectionTypeOptions.homevideos,
                    CollectionTypeOptions.mixed
                },
                GoreeCloudVideoLibraryPolicy.SupportedCollectionTypes);
        }

        [Theory]
        [InlineData(CollectionTypeOptions.movies)]
        [InlineData(CollectionTypeOptions.tvshows)]
        [InlineData(CollectionTypeOptions.homevideos)]
        [InlineData(CollectionTypeOptions.mixed)]
        public void IsSupported_VideoCollectionType_ReturnsTrue(CollectionTypeOptions collectionType)
        {
            Assert.True(GoreeCloudVideoLibraryPolicy.IsSupported(collectionType));
        }

        [Theory]
        [InlineData(CollectionTypeOptions.music)]
        [InlineData(CollectionTypeOptions.musicvideos)]
        [InlineData(CollectionTypeOptions.boxsets)]
        [InlineData(CollectionTypeOptions.books)]
        public void IsSupported_NonVideoFirstClassCollectionType_ReturnsFalse(CollectionTypeOptions collectionType)
        {
            Assert.False(GoreeCloudVideoLibraryPolicy.IsSupported(collectionType));
        }

        [Fact]
        public void IsSupported_UnspecifiedCollectionType_ReturnsFalse()
        {
            Assert.False(GoreeCloudVideoLibraryPolicy.IsSupported(null));
        }

        [Theory]
        [InlineData(CollectionTypeOptions.music)]
        [InlineData(CollectionTypeOptions.musicvideos)]
        [InlineData(CollectionTypeOptions.boxsets)]
        [InlineData(CollectionTypeOptions.books)]
        public void EnsureSupported_UnsupportedCollectionType_ThrowsArgumentException(CollectionTypeOptions collectionType)
        {
            var exception = Assert.Throws<ArgumentException>(() => GoreeCloudVideoLibraryPolicy.EnsureSupported(collectionType));

            Assert.Equal("collectionType", exception.ParamName);
        }

        [Fact]
        public void EnsureSupported_UnspecifiedCollectionType_ThrowsArgumentException()
        {
            var exception = Assert.Throws<ArgumentException>(() => GoreeCloudVideoLibraryPolicy.EnsureSupported(null));

            Assert.Equal("collectionType", exception.ParamName);
        }
    }
}
