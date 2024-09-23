import ink.francesco.stash.StashLibrary;
import ink.francesco.stash.Identity;
import ink.francesco.stash.DB;
import ink.francesco.stash.Stash;

public class TestLibrary {
    public static void main(String[] args) throws Exception {
        StashLibrary lib = StashLibrary.instance;
        System.out.println("Library loaded: " + lib);

        try {
            lib.stash_setLogLevel("info");
        } catch (Exception e) {
            System.out.println("Error setting log level: " + e);
        }

        var alice = new Identity("Alice");
        var bob = new Identity("Bob");

        DB db = DB.defaultDB();

        String url = String.format("file:///tmp/%s/sample", alice.id);
        Stash s = Stash.create(db, alice, url, new Stash.Config());
        s.close();

    }
    
}
