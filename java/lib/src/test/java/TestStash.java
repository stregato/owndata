
import java.io.IOException;
import java.util.List;

import org.junit.Test;

import ink.francesco.stash.Messanger;
import ink.francesco.stash.DB;
import ink.francesco.stash.Database;
import ink.francesco.stash.FS;
import ink.francesco.stash.Identity;
import ink.francesco.stash.Safe;
import ink.francesco.stash.StashConfig;
import ink.francesco.stash.StashLibrary;
import ink.francesco.stash.Transaction;

public class TestStash {

    @Test
    public void testStash() throws Exception {
       // StashConfig.libDir =  "./lib";
        StashLibrary.instance.stash_setLogLevel("info");

        Identity i = new Identity("Alice");

        DB db = DB.defaultDB();

        String url = String.format("file:///tmp/%s/sample", i.id);
        Safe s = Safe.create(db, i, url, new Safe.Config());
        s.close();
    }

    @Test
    public void testFS() throws Exception {
      //  StashConfig.libDir =  "./lib";
        StashLibrary.instance.stash_setLogLevel("info");

        Identity i = new Identity("Alice");

        DB db = DB.defaultDB();

        String url = String.format("file:///tmp/%s/sample", i.id);
        Safe s = Safe.create(db, i, url, new Safe.Config());
        FS fs = s.openFS();

        FS.File file = fs.putData("hello", "hello".getBytes(), new FS.PutOptions());
        assert file != null;
        assert file.id != 0;

        List<FS.File> files = fs.list("", new FS.ListOptions());
        assert files.size() == 1;
        assert files.get(0).id == file.id;
        assert files.get(0).name.equals(file.name);

        file = fs.putData("sub/world", "world".getBytes(), new FS.PutOptions());
        assert file != null;
        assert file.id != 0;
        assert file.name.equals("world");

        files = fs.list("sub", new FS.ListOptions());
        assert files.size() == 1;
        assert files.get(0).id == file.id;
        assert files.get(0).name.equals(file.name);

        fs.close();
        s.close();
    }

    @Test
    public void testDatabase() throws Exception {
        StashConfig.libDir =  "./lib";
        StashLibrary.instance.stash_setLogLevel("info");

        Identity i = new Identity("Alice");

        DB db = DB.defaultDB();

        String url = String.format("file:///tmp/%s/sample", i.id);
        Safe s = Safe.create(db, i, url, new Safe.Config());
        Database d = s.openDatabase(Safe.usrGroup, null);

        Transaction t = d.transaction();
        t.exec("CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY, name TEXT)", null);
        t.exec("INSERT INTO test (name) VALUES ('hello')", null);
        t.commit();
        
        d.sync();
        Database.Rows rows = d.query("SELECT * FROM test", null);
        while (true) {
            try {
                System.out.println(rows.next());
            } catch (Exception e) {
                break;
            }
        }
        rows.close();
        d.close();
        s.close();
    }

    @Test
    public void testMessanger() throws Exception {
        StashConfig.libDir =  "./lib";
        StashLibrary.instance.stash_setLogLevel("info");

        Identity i = new Identity("Alice");

        DB db = DB.defaultDB();

        String url = String.format("file:///tmp/%s/sample", i.id);
        Safe s = Safe.create(db, i, url, new Safe.Config());
        Messanger c = s.openMessanger();

        var m = new Messanger.Message();
        m.text = "hello";
        c.broadcast(Safe.usrGroup, m);

        List<Messanger.Message> messages = c.receive("");
        assert messages.size() == 1;
        assert messages.get(0).text.equals("hello");

        s.close();
    }

    public static void main(String[] args) throws IOException {
        System.setProperty("jna.debug_load", "true");
        System.setProperty("jna.debug_load.jna", "true");
        
        String currentDir = System.getProperty("user.dir");
        System.out.println("Current directory: " + currentDir ); 
        StashConfig.libDir =  currentDir+"/java/lib";

        Identity identity = new Identity("Alice");
        System.out.println(identity.toJson());
    }
}
