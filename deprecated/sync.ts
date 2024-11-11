import { mkdtemp } from "fs/promises";
import { tmpdir } from "os";
import { promises as fs } from "fs";
import { join } from "path";
import git from "isomorphic-git";
import http from "isomorphic-git/http/node";

function getMessage(msg: string): string {
  return JSON.stringify({ message: msg });
}

async function sync(
  upstreamRepo: string,
  upstreamBranch: string,
  forkRepo: string,
  forkBranch: string,
  pat: string
) {
  const upstreamUrl = `https://github.com/${upstreamRepo}.git`;
  const forkUrl = `https://github.com/${forkRepo}.git`;

  const tempDir = await mkdtemp(join(tmpdir(), "repo-"));

  await git.clone({
    fs,
    http,
    dir: tempDir,
    url: upstreamUrl,
    singleBranch: true,
    ref: upstreamBranch,
    onAuth: () => ({ username: pat }),
  });

  // Add the fork as a remote
  await git.addRemote({
    fs,
    dir: tempDir,
    remote: "fork",
    url: forkUrl,
  });

  // Push to the fork branch
  await git.push({
    fs,
    http,
    dir: tempDir,
    remote: "fork",
    ref: upstreamBranch,
    remoteRef: forkBranch,
    force: true,
    onAuth: () => ({ username: pat }),
  });
}

export async function POST(request: Request): Promise<Response> {
  const params = new URLSearchParams(request.url.split("?")[1]);
  const forkRepo = params.get("fork_repo");
  const forkBranch = params.get("fork_branch");
  let upstreamRepo = params.get("upstream_repo");
  let upstreamBranch = params.get("upstream_branch");
  const pat = params.get("pat");

  // if (upstreamRepo === null || upstreamBranch === null) {
  //   const body = await request.json();
  //   const { repository, ref } = body;
  //   if (!repository || !ref) {
  //     return new Response(
  //       getMessage("No repository or ref found in the payload"),
  //       {
  //         status: 400,
  //       }
  //     );
  //   }

  //   upstreamRepo = repository.full_name;
  //   upstreamBranch = ref.replace("refs/heads/", "");
  //   if (!upstreamRepo) {
  //     return new Response(getMessage("No repository found in the payload"), {
  //       status: 400,
  //     });
  //   }
  // }

  if (!forkRepo || !forkBranch || !upstreamRepo || !upstreamBranch || !pat) {
    return new Response(
      getMessage(
        "Please provide fork_repo, fork_branch, upstream_repo, upstream_branch and pat"
      ),
      { status: 400 }
    );
  }

  try {
    await sync(upstreamRepo, upstreamBranch, forkRepo, forkBranch, pat);
  } catch (error) {
    return new Response(getMessage(error.message), {
      status: 500,
    });
  }

  return new Response(getMessage("OK"));
}
