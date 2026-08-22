using System.Collections.Generic;
using MediaBrowser.Common.Api;
using MediaBrowser.Controller.Library;
using MediaBrowser.Model.Entities;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;

namespace Jellyfin.Api.Controllers;

/// <summary>
/// Exposes GoreeCloud-owned video library capabilities for first-party clients.
/// </summary>
[Route("GoreeCloud/Video/Libraries")]
[Authorize(Policy = Policies.FirstTimeSetupOrElevated)]
public sealed class GoreeCloudVideoLibraryController : BaseJellyfinApiController
{
    /// <summary>
    /// Gets the collection types that can be created as first-class GoreeCloud Video libraries.
    /// </summary>
    /// <response code="200">Supported collection types returned.</response>
    /// <returns>The supported inherited collection types.</returns>
    [HttpGet("SupportedTypes")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    public ActionResult<IReadOnlyList<CollectionTypeOptions>> GetSupportedCollectionTypes()
    {
        return Ok(GoreeCloudVideoLibraryPolicy.SupportedCollectionTypes);
    }
}
