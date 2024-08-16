import java.nio.file.Path;
import java.nio.file.Paths;

public class DB {
    long hnd;

    static public DB defaultDB() throws Exception {
                String osName = System.getProperty("os.name").toLowerCase();
        String userHome = System.getProperty("user.home");
        Path configFolder;

        if (osName.contains("win")) {
            // Windows: Use APPDATA
            String appData = System.getenv("APPDATA");
            configFolder = Paths.get(appData, "mio.db");
        } else if (osName.contains("mac")) {
            // macOS: Use Library/Application Support
            configFolder = Paths.get(userHome, "Library", "Application Support", "mio.db");
        } else {
            // Linux/Unix: Use .config directory
            configFolder = Paths.get(userHome, ".config", "mio.db");
        }
        return new DB(configFolder.toString());
    }

    public DB(String path) throws Exception {
        Result r = StashLibrary.instance.stash_openDB(path);
        r.check();
        hnd = r.hnd;

    }

    public void close() {
        StashLibrary.instance.stash_closeDB(hnd);
    }
}
