using System;
using System.Collections.Generic;
using MediaBrowser.Model.Entities;

namespace MediaBrowser.Controller.Library
{
    /// <summary>
    /// Defines the first-class library collection types supported by GoreeCloud Video.
    /// </summary>
    /// <remarks>
    /// This policy intentionally preserves Jellyfin's inherited collection-type contract while
    /// preventing non-video library types from being created through GoreeCloud Video's library
    /// management surface. Shared audio, subtitle, artwork, probing, and transcoding capabilities
    /// remain available where they are required for video playback.
    /// </remarks>
    public static class GoreeCloudVideoLibraryPolicy
    {
        private static readonly IReadOnlyList<CollectionTypeOptions> _supportedCollectionTypes = Array.AsReadOnly(
            new[]
            {
                CollectionTypeOptions.movies,
                CollectionTypeOptions.tvshows,
                CollectionTypeOptions.homevideos,
                CollectionTypeOptions.mixed
            });

        /// <summary>
        /// Gets the inherited collection types exposed as first-class GoreeCloud Video libraries.
        /// </summary>
        public static IReadOnlyList<CollectionTypeOptions> SupportedCollectionTypes => _supportedCollectionTypes;

        /// <summary>
        /// Determines whether an inherited collection type is supported as a first-class
        /// GoreeCloud Video library.
        /// </summary>
        /// <param name="collectionType">The inherited collection type.</param>
        /// <returns><see langword="true"/> when the collection type is video-oriented.</returns>
        public static bool IsSupported(CollectionTypeOptions? collectionType)
        {
            return collectionType is CollectionTypeOptions.movies
                or CollectionTypeOptions.tvshows
                or CollectionTypeOptions.homevideos
                or CollectionTypeOptions.mixed;
        }

        /// <summary>
        /// Ensures that a requested library collection type is supported by GoreeCloud Video.
        /// </summary>
        /// <param name="collectionType">The requested inherited collection type.</param>
        /// <exception cref="ArgumentException">Thrown when the requested type is not supported.</exception>
        public static void EnsureSupported(CollectionTypeOptions? collectionType)
        {
            if (IsSupported(collectionType))
            {
                return;
            }

            var requestedType = collectionType?.ToString() ?? "unspecified";
            throw new ArgumentException(
                $"Collection type '{requestedType}' is not supported by GoreeCloud Video. Supported types are movies, tvshows, homevideos, and mixed.",
                nameof(collectionType));
        }
    }
}
