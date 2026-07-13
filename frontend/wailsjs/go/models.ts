export namespace domains {
	
	export class ChangePasswordDTO {
	    newPassword: string;
	    oldPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new ChangePasswordDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.newPassword = source["newPassword"];
	        this.oldPassword = source["oldPassword"];
	    }
	}
	export class Reaction {
	    id: number;
	    emoji: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new Reaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.emoji = source["emoji"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class MessageReactionSummary {
	    reaction: Reaction;
	    userIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new MessageReactionSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reaction = this.convertValues(source["reaction"], Reaction);
	        this.userIds = source["userIds"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Message {
	    id: number;
	    chatId: number;
	    sender: User;
	    text: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    isRead: boolean;
	    replyToMessage?: Message;
	    reactions?: MessageReactionSummary[];
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.chatId = source["chatId"];
	        this.sender = this.convertValues(source["sender"], User);
	        this.text = source["text"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.isRead = source["isRead"];
	        this.replyToMessage = this.convertValues(source["replyToMessage"], Message);
	        this.reactions = this.convertValues(source["reactions"], MessageReactionSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class User {
	    id: number;
	    username: string;
	    email: string;
	    emailConfirmed: boolean;
	    password: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    avatarPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.email = source["email"];
	        this.emailConfirmed = source["emailConfirmed"];
	        this.password = source["password"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.avatarPath = source["avatarPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Chat {
	    id: number;
	    title?: string;
	    description?: string;
	    type: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    unreadCount: number;
	    members?: User[];
	    messages?: Message[];
	
	    static createFrom(source: any = {}) {
	        return new Chat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.type = source["type"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.unreadCount = source["unreadCount"];
	        this.members = this.convertValues(source["members"], User);
	        this.messages = this.convertValues(source["messages"], Message);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CreateChatDTO {
	    title?: string;
	    description?: string;
	    type: string;
	    memberIDs?: number[];
	
	    static createFrom(source: any = {}) {
	        return new CreateChatDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.type = source["type"];
	        this.memberIDs = source["memberIDs"];
	    }
	}
	export class ForgetPasswordDTO {
	    newPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new ForgetPasswordDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.newPassword = source["newPassword"];
	    }
	}
	export class LoginDTO {
	    login: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.login = source["login"];
	        this.password = source["password"];
	    }
	}
	
	
	export class Pagination {
	    limit?: number;
	    offset?: number;
	
	    static createFrom(source: any = {}) {
	        return new Pagination(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.limit = source["limit"];
	        this.offset = source["offset"];
	    }
	}
	
	export class RegisterDTO {
	    username: string;
	    email: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new RegisterDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.email = source["email"];
	        this.password = source["password"];
	    }
	}
	export class Settings {
	    theme: number;
	    emailConsents: number;
	    webPushConsents: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.emailConsents = source["emailConsents"];
	        this.webPushConsents = source["webPushConsents"];
	    }
	}
	export class UpdateUserDTO {
	    username?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateUserDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	    }
	}
	
	export class UsersFilters {
	    username?: string;
	
	    static createFrom(source: any = {}) {
	        return new UsersFilters(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	    }
	}

}

