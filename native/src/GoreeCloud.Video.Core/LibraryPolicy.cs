namespace GoreeCloud.Video.Core;

public enum LibraryKind
{
    Movies,
    Television,
    HomeVideos,
    FamilyMedia,
    OtherVideos,
}

/// <summary>
/// First-party GoreeCloud Video library policy. This clean native boundary deliberately
/// models only video product concepts and does not inherit Jellyfin collection-type enums.
/// </summary>
public static class LibraryPolicy
{
    public static bool IsSupported(LibraryKind kind) => kind switch
    {
        LibraryKind.Movies => true,
        LibraryKind.Television => true,
        LibraryKind.HomeVideos => true,
        LibraryKind.FamilyMedia => true,
        LibraryKind.OtherVideos => true,
        _ => false,
    };
}

public sealed record VideoLibrary(
    Guid Id,
    Guid OwnerId,
    string Name,
    LibraryKind Kind,
    string RootPath,
    DateTimeOffset CreatedAt);

public interface IVideoLibraryRepository
{
    Task<IReadOnlyList<VideoLibrary>> ListAsync(Guid ownerId, CancellationToken cancellationToken);
    Task<VideoLibrary?> GetAsync(Guid ownerId, Guid libraryId, CancellationToken cancellationToken);
}
