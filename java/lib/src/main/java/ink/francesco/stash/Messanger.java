package ink.francesco.stash;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.util.List;

public class Messanger {

    static public class Message {
        public String sender;
        public int encryptionId;
        public String recipient;
        public long id;
        public String text;
        public byte[] data;
        public String file;
    }


    private final long hnd;
    private final ObjectMapper mapper = new ObjectMapper();

    public Messanger(long hnd) {
        this.hnd = hnd;
    }

    public void send(String userId, Message message) throws Exception {
        Result r = StashLibrary.instance.stash_send(hnd, userId, mapper.writeValueAsString(message));
        r.check();
    }

    public void broadcast(String groupName, Message message) throws Exception {
        Result r = StashLibrary.instance.stash_broadcast(hnd, groupName, mapper.writeValueAsString(message));
        r.check();
    }

    public List<Message> receive(String filter) throws Exception {
        Result r = StashLibrary.instance.stash_receive(hnd, filter);
        return r.list(Message.class);
    }

    public void download(Message message, String dest) throws Exception {
        Result r = StashLibrary.instance.stash_download(hnd, mapper.writeValueAsString(message), dest);
        r.check();
    }

    public void rewind(String dest, long messageId) throws Exception {
        Result r = StashLibrary.instance.stash_rewind(hnd, dest, messageId);
        r.check();
    }

}
