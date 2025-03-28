import SyncList from './SyncList';

export default function GithubSync() {
    return (
        <div id="github-sync">
            <a href="/">Home</a>
            <div id="new-sync">

            </div>
            <div id="current-syncs">
                <SyncList />
            </div>
        </div>
    )
}
