package ink.francesco.stash;

import java.io.File;
import java.net.URL;

import com.sun.jna.Library;
import com.sun.jna.Native;
import com.sun.jna.Platform;



public interface StashLibrary extends Library {

    // stash_setLogLevel sets the log level of the mio library.
    Result stash_setLogLevel(String level);

    // stash_newIdentity creates a new identity with the given nick.
    Result stash_newIdentity(String nick);

    // stash_openDB opens a database at the given file path
    Result stash_openDB(String path);

    // stash_closeDB closes the database with the given handle
    Result stash_closeDB(long dbH);

    // stash_createSafe creates a new safe with the given identity, url and config
    Result stash_createSafe(long dbH, String identity, String url, String config);

    // stash_openSafe opens a safe with the given identity and url
    Result stash_openSafe(long dbH, String identity, String url);

    // stash_closeStash closes the safe with the given handle
    Result stash_closeSafe(long safeH);
    
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
        String libExtension = "";
        String osFolder = "";
        String arch = System.getProperty("os.arch");
    
        // Determine the platform-specific folder and library extension
        if (Platform.isWindows()) {
            osFolder = "windows_" + arch;
            libExtension = ".dll";
        } else if (Platform.isMac()) {
            osFolder = arch.contains("aarch") ? "darwin_arm64" : "darwin_amd64";
            libName = "lib" + libName; // Add "lib" prefix for Unix-like systems
            libExtension = ".dylib";
        } else if (Platform.isLinux()) {
            osFolder = "linux_" + arch;
            libName = "lib" + libName; // Add "lib" prefix for Unix-like systems
            libExtension = ".so";
        } else if (Platform.isAndroid()) {
            osFolder = "android_arm64";
            libName = "lib" + libName; // Add "lib" prefix for Unix-like systems
            libExtension = ".so";
        }
    
        // Construct the library name for JNA to load
        String resourceLibName = libName + libExtension;
    
        StashLibrary libraryInstance = null;
    
        try {
            // First attempt to load the library via JNA (without specifying full path)
            libraryInstance = (StashLibrary) Native.load(libName, StashLibrary.class);
        } catch (UnsatisfiedLinkError e) {
            // Calculate fallback path relative to the location of the class
            URL classLocation = StashLibrary.class.getProtectionDomain().getCodeSource().getLocation();
            File classFile = new File(classLocation.getPath());
            File classDir = classFile.isDirectory() ? classFile : classFile.getParentFile();
    
            // Now calculate the path to the build directory relative to the class location
            File buildDir = new File(classDir, "../../../../build/" + osFolder + "/" + resourceLibName);
            if (buildDir.exists()) {
                // Load the library using the full path
                libraryInstance = (StashLibrary) Native.load(buildDir.getAbsolutePath(), StashLibrary.class);
            } else {
                throw new RuntimeException("Failed to load native library from both JAR and ../build directory", e);
            }
        }
    
        // Synchronize the loaded library
        return (StashLibrary) Native.synchronizedLibrary(libraryInstance);
    }

    public StashLibrary instance = StashLibrary.loadLibrary();
}
