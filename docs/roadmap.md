# Roadmap

`sake` is under active development. Before **v1.0.0**, I want to finish the following tasks, some miscellaneous fixes, improve code documentation and refactor:

- [ ] Task not callable, only from another task (as not to accidently call it)
- [ ] Hide tasks from auto-completion via `hidden: true` attribute
- [ ] Silent output from task via `silent: true` (and flag)
- [ ] ExecTTY should be config shell
- [ ] Add flag `default_timeout_s`
- [ ] Use `chdir` for tasks, `work_dir` for servers
- [ ] Move limit/limitp to spec, or move order to target
- [ ] Figure out changed/skipped/when
- [ ] Conditional tasks (success, error, skip)
- [ ] Add callbacks (success/error)
- [ ] Loader show current task and how many left on table
- [ ] Add retries to task
- [ ] Add required envs
- [ ] Add option to prompt for envs
- [ ] Handle `Match *` in ssh config for inventory as well
- [ ] Something similar to play, to trigger multiple tasks (with their own context)
- [ ] Add env variables to multiple servers
- [ ] Run one task, save output from all, and then have one task handle differences
- [ ] Save logs/output to files (remote/local)
- [ ] Diff task
- [ ] Inherit default from `default` spec/target
- [ ] Add yaml to command mapper
- [ ] Implement facts
- [ ] Configure what to show, host/ip or name, configure via theme flags
   - [x] Template for server prefix, similar to header
   - [ ] Add colors to describe (key bold, value color), true (green), false (red)
   - [ ] Add Tree output
- [ ] Fix hashed ip6 with port 22 does not work, all other combinations work
- [ ] Fix `sake ssh inv` not working
