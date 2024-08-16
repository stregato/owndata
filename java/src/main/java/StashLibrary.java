import java.io.File;

import com.sun.jna.Library;
import com.sun.jna.Native;
import com.sun.jna.Platform;

class StashConfig {
    public static String libDir = "";
}

public interface StashLibrary extends Library {

    // stash_setLogLevel sets the log level of the mio library.
    Result stash_setLogLevel(String level);

    // stash_newIdentity creates a new identity with the given nick.
    Result stash_newIdentity(String nick);

    // stash_openDB opens a database at the given file path
    Result stash_openDB(String path);

    // stash_closeDB closes the database with the given handle
    Result stash_closeDB(long dbH);

    // stash_createStash creates a new safe with the given identity, url and config
    Result stash_createStash(long dbH, String identity, String url, String config);

    // stash_openStash opens a safe with the given identity and url
    Result stash_openStash(long dbH, String identity, String url);

    // stash_closeStash closes the safe with the given handle
    Result stash_closeStash(long safeH);
    
    // stash_createGroup creates a new group with the given name
    Result stash_updateGroup(long safeH, String groupName, long change, String users);

    Result stash_getGroups(long safeH);

    Result stash_getKeys(long safeH, String groupName, long expectedMinimumLenght);

    Result stash_openFS(long safeH);

    Result stash_closeFS(long fsH);

    Result stash_list(long fsH, String path, String options);

    Result stash_stat(long fsH, String path);

    Result stash_putFile(long fsH, String dest, String src, String options);

    Result stash_putData(long fsH, String dest, Data data, String options);

    Result stash_getFile(long fsH, String src, String dest, String options);

    Result stash_getData(long fsH, String src, String options);

    Result stash_delete(long fsH, String path);

    Result stash_rename(long fsH, String oldPath, String newPath);

    Result stash_openDatabase(long safeH, String groupName, String ddls);

    Result stash_closeDatabase(long dbH);

    Result stash_exec(long dbH, String query, String args);

    Result stash_query(long dbH, String key, String args);

    Result stash_nextRow(long rowsH);

    Result stash_closeRows(long rowsH);

    Result stash_sync(long dbH);

    Result stash_cancel(long dbH);

    Result stash_openComm(long safeH);

    Result stash_rewind(long commH, String dest, long messageID);

    Result stash_send(long commH, String userId, String message);

    Result stash_broadcast(long commH, String groupName, String message);

    Result stash_receive(long commH, String filter);

    Result stash_download(long commH, String message, String dest);
    
    static StashLibrary loadLibrary() {
        String libName = "stash";
        String libPath;

//        Native.setProtected(true);

        if (Platform.isWindows()) {
            libPath = libName + ".dll";
        } else if (Platform.isMac()) {
            libPath = "lib" + libName + ".dylib";
        } else {
            // Assume Linux or Unix-like
            libPath = "lib" + libName + ".so";
        }
        if (!StashConfig.libDir.isEmpty()) {
            libPath = StashConfig.libDir + "/" + libPath;
        }

        StashLibrary i;
        File libFile = new File(libPath);

        if (libFile.exists()) {
            i = (StashLibrary) Native.load(libFile.getAbsolutePath(), StashLibrary.class);
        } else {
            // If the file does not exist, try loading by name
            i = (StashLibrary) Native.load(libName, StashLibrary.class);
        }

        i = (StashLibrary) Native.synchronizedLibrary(i);
        return i;
    }

    public StashLibrary instance = StashLibrary.loadLibrary();
}
