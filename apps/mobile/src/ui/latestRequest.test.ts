import { createLatestRequestGuard } from "./latestRequest";

function deferred<Value>() {
  let resolve!: (value: Value) => void;
  const promise = new Promise<Value>((done) => { resolve = done; });
  return { promise, resolve };
}

describe("latest request guard", () => {
  it("keeps rapid Teacher A → B loading from committing stale A Lessons", async () => {
    const guard = createLatestRequestGuard();
    const teacherA = deferred<string[]>();
    const teacherB = deferred<string[]>();
    const commits: string[][] = [];
    const load = async (promise: Promise<string[]>) => {
      const request = guard.begin();
      const result = await promise;
      if (request.isCurrent()) commits.push(result);
      return request.signal.aborted;
    };

    const first = load(teacherA.promise);
    const second = load(teacherB.promise);
    teacherB.resolve(["lesson_b"]);
    await second;
    teacherA.resolve(["lesson_a"]);

    await expect(first).resolves.toBe(true);
    expect(commits).toEqual([["lesson_b"]]);
  });
});
